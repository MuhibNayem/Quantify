package services

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
	"inventory/backend/internal/units"
)

const discountAuditThresholdPct = 10.0

type SalesService struct {
	DB               *gorm.DB
	Settings         SettingsService
	ReportingService *ReportingService
}

func NewSalesService(db *gorm.DB, settings SettingsService, reportingService *ReportingService) *SalesService {
	return &SalesService{
		DB:               db,
		Settings:         settings,
		ReportingService: reportingService,
	}
}

func (s *SalesService) ProcessCheckout(req requests.CheckoutRequest, userID uint) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		products, err := s.fetchProducts(tx, req.Items)
		if err != nil {
			return err
		}

		batchesByProduct, err := s.fetchBatches(tx, req.Items)
		if err != nil {
			return err
		}

		activePromotions, err := s.fetchActivePromotions(tx)
		if err != nil {
			return err
		}

		subtotal, totalDiscountFromPromotions, orderItems, err := s.processItems(tx, req, products, batchesByProduct, activePromotions, userID)
		if err != nil {
			return err
		}

		postTaxTotal, taxAmount, err := s.applyTax(subtotal)
		if err != nil {
			return err
		}

		discountAmount, pointsRedeemed, err := s.handleLoyalty(tx, req.CustomerID, req.PointsToRedeem, postTaxTotal)
		if err != nil {
			return err
		}

		return s.createOrderAndTransaction(tx, req, userID, subtotal, taxAmount, discountAmount, totalDiscountFromPromotions, pointsRedeemed, orderItems)
	})
}

func (s *SalesService) fetchActivePromotions(tx *gorm.DB) ([]domain.Promotion, error) {
	var activePromotions []domain.Promotion
	now := time.Now()
	if err := tx.Preload("Product").Preload("Category").Preload("SubCategory").
		Where("is_active = ? AND start_date <= ? AND end_date >= ?", true, now, now).
		Find(&activePromotions).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch promotions: %w", err)
	}
	return activePromotions, nil
}

func (s *SalesService) processItems(
	tx *gorm.DB,
	req requests.CheckoutRequest,
	productMap map[uint]domain.Product,
	batchesByProduct map[uint][]*domain.Batch,
	activePromotions []domain.Promotion,
	userID uint,
) (float64, float64, []domain.OrderItem, error) {
	var totalAmount float64
	var totalDiscountFromPromotions float64
	var orderItems []domain.OrderItem

	for _, item := range req.Items {
		product, ok := productMap[item.ProductID]
		if !ok {
			return 0, 0, nil, fmt.Errorf("product %d not found", item.ProductID)
		}

		if !product.AllowFractional && item.Quantity != math.Trunc(item.Quantity) {
			return 0, 0, nil, fmt.Errorf("product %q cannot be sold in fractional quantities", product.Name)
		}

		// Resolve selling unit + conversion factor (default: base unit, factor 1).
		sellingFactor := 1.0
		var unitID *uint
		var unitPriceOverride float64 // per selling unit, 0 = derive
		if item.UnitID != nil {
			pu, err := s.resolveProductUnit(tx, product, *item.UnitID)
			if err != nil {
				return 0, 0, nil, err
			}
			sellingFactor = pu.Factor
			unitID = item.UnitID
			unitPriceOverride = pu.Price
		}

		// Convert selling quantity → base stock quantity, deduct stock in BASE units.
		baseQty := units.StockDeduction(item.Quantity, sellingFactor)

		deductionLocationID := resolveDeductionLocation(product, req.RegisterLocationID)

		var batches []*domain.Batch
		if deductionLocationID != 0 {
			local, err := s.fetchBatchesForLocation(tx, product.ID, deductionLocationID)
			if err != nil {
				return 0, 0, nil, err
			}
			batches = local
		}
		if len(batches) == 0 {
			batches = batchesByProduct[item.ProductID]
		}

		if _, err := s.allocateStock(tx, product, batches, deductionLocationID, baseQty, req.CustomerID, userID); err != nil {
			return 0, 0, nil, err
		}

		// Price per SELLING unit: explicit ProductUnit.Price wins, else base price × factor.
		baseUnitPrice := units.EffectivePrice(unitPriceOverride, product.SellingPrice, sellingFactor)
		unitPrice, discount := s.applyPromotion(baseUnitPrice, product, item.Quantity, activePromotions)
		totalDiscountFromPromotions += discount
		totalAmount += unitPrice * item.Quantity

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:    item.ProductID,
			Quantity:     item.Quantity, // selling units (e.g. 14 inches)
			UnitID:       unitID,
			BaseQuantity: baseQty, // base units deducted (e.g. 1.1667 ft)
			UnitPrice:    unitPrice,
			TotalPrice:   units.RoundQty(unitPrice * item.Quantity),
		})
	}

	return totalAmount, totalDiscountFromPromotions, orderItems, nil
}

func (s *SalesService) createOrderAndTransaction(
	tx *gorm.DB,
	req requests.CheckoutRequest,
	userID uint,
	subtotal float64,
	taxAmount float64,
	discountAmount float64,
	totalDiscountFromPromotions float64,
	pointsRedeemed int,
	orderItems []domain.OrderItem,
) error {
	orderNumber := fmt.Sprintf("ORD-%d-%d", time.Now().Unix(), userID)
	finalTotal := subtotal + taxAmount - discountAmount
	order := domain.Order{
		OrderNumber:    orderNumber,
		UserID:         userID,
		CustomerID:     req.CustomerID,
		Subtotal:       subtotal,
		TaxAmount:      taxAmount,
		TotalAmount:    finalTotal,
		Status:         "COMPLETED",
		PaymentMethod:  req.PaymentMethod,
		OrderDate:      time.Now(),
		PointsRedeemed: pointsRedeemed,
		DiscountAmount: discountAmount + totalDiscountFromPromotions,
	}

	if err := tx.Create(&order).Error; err != nil {
		return fmt.Errorf("failed to create order record: %w", err)
	}

	for i := range orderItems {
		orderItems[i].OrderID = order.ID
	}
	if err := tx.Create(&orderItems).Error; err != nil {
		return fmt.Errorf("failed to create order items: %w", err)
	}

	saletransaction := domain.Transaction{
		OrderID:              orderNumber,
		Amount:               int64(finalTotal * 100),
		Currency:             "USD",
		PaymentMethod:        req.PaymentMethod,
		Status:               "COMPLETED",
		GatewayTransactionID: fmt.Sprintf("GW-%d", time.Now().UnixNano()),
	}

	if err := tx.Create(&saletransaction).Error; err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	s.emitDiscountAuditIfSignificant(tx, orderNumber, userID, subtotal, discountAmount, totalDiscountFromPromotions)

	return nil
}

func (s *SalesService) emitDiscountAuditIfSignificant(
	tx *gorm.DB,
	orderNumber string,
	userID uint,
	subtotalAfterPromo float64,
	loyaltyDiscount float64,
	promotionDiscount float64,
) {
	preDiscountSubtotal := subtotalAfterPromo + promotionDiscount
	if preDiscountSubtotal <= 0 {
		return
	}
	totalDiscount := loyaltyDiscount + promotionDiscount
	discountPct := (totalDiscount / preDiscountSubtotal) * 100.0
	if discountPct < discountAuditThresholdPct {
		return
	}

	changes, err := json.Marshal(map[string]float64{
		"subtotal":           round2(preDiscountSubtotal),
		"total_discount":     round2(totalDiscount),
		"discount_pct":       round2(discountPct),
		"promotion_discount": round2(promotionDiscount),
		"loyalty_discount":   round2(loyaltyDiscount),
	})
	if err != nil {
		changes = []byte(`{}`)
	}

	entry := &domain.AuditLog{
		Action:    "DISCOUNT",
		Entity:    "Order",
		EntityID:  orderNumber,
		UserID:    userID,
		Changes:   string(changes),
		Timestamp: time.Now(),
	}
	_ = repository.WriteAudit(tx.Statement.Context, tx, entry)
}
