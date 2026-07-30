package services

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
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

		subtotal, totalDiscountFromPromotions, cogs, orderItems, err := s.processItems(tx, req, products, batchesByProduct, activePromotions, userID)
		if err != nil {
			return err
		}

		// Manual discount (admin opt-in) reduces the subtotal before tax.
		manualDiscountAmount, err := s.computeManualDiscount(subtotal, req)
		if err != nil {
			return err
		}
		discountedSubtotal := round2(subtotal - manualDiscountAmount)

		postTaxTotal, taxAmount, err := s.applyTax(discountedSubtotal)
		if err != nil {
			return err
		}

		discountAmount, pointsRedeemed, err := s.handleLoyalty(tx, req.CustomerID, req.PointsToRedeem, postTaxTotal)
		if err != nil {
			return err
		}

		return s.createOrderAndTransaction(tx, req, userID, subtotal, discountedSubtotal, taxAmount, discountAmount, totalDiscountFromPromotions, manualDiscountAmount, pointsRedeemed, cogs, orderItems)
	})
}

// computeManualDiscount validates an employee's manual discount against the
// admin-configured policy and returns the currency amount to deduct.
func (s *SalesService) computeManualDiscount(subtotal float64, req requests.CheckoutRequest) (float64, error) {
	if req.ManualDiscountType == "" || req.ManualDiscountValue <= 0 {
		return 0, nil
	}
	// Master switch.
	if !s.settingBool("manual_discount_enabled", false) {
		return 0, fmt.Errorf("manual discounts are disabled by policy")
	}
	// Reason required?
	if s.settingBool("manual_discount_require_reason", true) && req.ManualDiscountReason == "" {
		return 0, fmt.Errorf("a reason is required for manual discounts")
	}

	var amount float64
	switch req.ManualDiscountType {
	case "percent":
		maxPct := s.settingFloat("manual_discount_max_percent", 10)
		if req.ManualDiscountValue > maxPct {
			return 0, fmt.Errorf("discount %.2f%% exceeds the maximum allowed %.2f%%", req.ManualDiscountValue, maxPct)
		}
		amount = subtotal * (req.ManualDiscountValue / 100.0)
	case "fixed":
		maxAmt := s.settingFloat("manual_discount_max_amount", 0)
		if maxAmt > 0 && req.ManualDiscountValue > maxAmt {
			return 0, fmt.Errorf("discount %.2f exceeds the maximum allowed %.2f", req.ManualDiscountValue, maxAmt)
		}
		amount = req.ManualDiscountValue
	default:
		return 0, fmt.Errorf("invalid manual discount type")
	}
	if amount > subtotal {
		amount = subtotal // never discount more than the subtotal
	}
	return round2(amount), nil
}

// settingBool reads a boolean setting, falling back to def on error/parse failure.
func (s *SalesService) settingBool(key string, def bool) bool {
	v, err := s.Settings.GetSetting(key)
	if err != nil {
		return def
	}
	return v == "true" || v == "1"
}

// settingFloat reads a numeric setting, falling back to def on error/parse failure.
func (s *SalesService) settingFloat(key string, def float64) float64 {
	v, err := s.Settings.GetSetting(key)
	if err != nil {
		return def
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return def
	}
	return f
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
) (float64, float64, float64, []domain.OrderItem, error) {
	var totalAmount float64
	var totalDiscountFromPromotions float64
	var totalCOGS float64
	var orderItems []domain.OrderItem

	for _, item := range req.Items {
		product, ok := productMap[item.ProductID]
		if !ok {
			return 0, 0, 0, nil, fmt.Errorf("product %d not found", item.ProductID)
		}

		if !product.AllowFractional && item.Quantity != math.Trunc(item.Quantity) {
			return 0, 0, 0, nil, fmt.Errorf("product %q cannot be sold in fractional quantities", product.Name)
		}

		// Resolve selling unit + conversion factor (default: base unit, factor 1).
		sellingFactor := 1.0
		var unitID *uint
		var unitPriceOverride float64 // per selling unit, 0 = derive
		if item.UnitID != nil {
			pu, err := s.resolveProductUnit(tx, product, *item.UnitID)
			if err != nil {
				return 0, 0, 0, nil, err
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
				return 0, 0, 0, nil, err
			}
			batches = local
		}
		if len(batches) == 0 {
			batches = batchesByProduct[item.ProductID]
		}

		if _, err := s.allocateStock(tx, product, batches, deductionLocationID, baseQty, req.CustomerID, userID); err != nil {
			return 0, 0, 0, nil, err
		}

		// Price per SELLING unit: explicit ProductUnit.Price wins, else base price × factor.
		baseUnitPrice := units.EffectivePrice(unitPriceOverride, product.SellingPrice, sellingFactor)
		unitPrice, discount := s.applyPromotion(baseUnitPrice, product, item.Quantity, activePromotions)
		totalDiscountFromPromotions += discount
		totalAmount += unitPrice * item.Quantity

		// Cost of goods sold: base quantity deducted × purchase price per base unit.
		totalCOGS += baseQty * product.PurchasePrice

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:    item.ProductID,
			Quantity:     item.Quantity, // selling units (e.g. 14 inches)
			UnitID:       unitID,
			BaseQuantity: baseQty, // base units deducted (e.g. 1.1667 ft)
			UnitPrice:    unitPrice,
			TotalPrice:   units.RoundQty(unitPrice * item.Quantity),
		})
	}

	return totalAmount, totalDiscountFromPromotions, totalCOGS, orderItems, nil
}

func (s *SalesService) createOrderAndTransaction(
	tx *gorm.DB,
	req requests.CheckoutRequest,
	userID uint,
	subtotal float64,
	discountedSubtotal float64,
	taxAmount float64,
	discountAmount float64,
	totalDiscountFromPromotions float64,
	manualDiscountAmount float64,
	pointsRedeemed int,
	cogs float64,
	orderItems []domain.OrderItem,
) error {
	orderNumber := fmt.Sprintf("ORD-%d-%d", time.Now().Unix(), userID)
	finalTotal := discountedSubtotal + taxAmount - discountAmount
	onCredit := req.PaymentMethod == "credit"
	// A credit (unpaid) sale stays PENDING until paid; other methods complete now.
	status := "COMPLETED"
	if onCredit {
		status = "PENDING"
	}
	order := domain.Order{
		OrderNumber:    orderNumber,
		UserID:         userID,
		CustomerID:     req.CustomerID,
		Subtotal:       discountedSubtotal,
		TaxAmount:      taxAmount,
		TotalAmount:    finalTotal,
		Status:         status,
		PaymentMethod:  req.PaymentMethod,
		OrderDate:      time.Now(),
		PointsRedeemed: pointsRedeemed,
		DiscountAmount: discountAmount + totalDiscountFromPromotions + manualDiscountAmount,
		ManualDiscountType:   req.ManualDiscountType,
		ManualDiscountValue:  req.ManualDiscountValue,
		ManualDiscountAmount: manualDiscountAmount,
		ManualDiscountReason: req.ManualDiscountReason,
		ManualDiscountBy:     userID,
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
		Status:               status,
		GatewayTransactionID: fmt.Sprintf("GW-%d", time.Now().UnixNano()),
	}

	if err := tx.Create(&saletransaction).Error; err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	// Post the double-entry ledger entry for the sale (DR Cash/AR, CR Revenue,
	// DR Sales Discounts, DR COGS, CR Inventory). Keeps income and cost balanced.
	ledger := NewLedgerService(s.DB)
	entry, err := ledger.PostSale(tx, &order, cogs, manualDiscountAmount, onCredit, userID)
	if err != nil {
		return fmt.Errorf("failed to post sale to ledger: %w", err)
	}

	// For an unpaid (credit) delivery, record the amount the customer owes.
	if onCredit {
		if req.CustomerID == nil {
			return fmt.Errorf("credit sales require a customer")
		}
		ar := domain.AccountsReceivable{
			CustomerID:     *req.CustomerID,
			OrderID:        order.ID,
			Amount:         decimal.NewFromFloat(finalTotal).Round(4),
			PaidAmount:     decimal.Zero,
			Status:         "UNPAID",
			DueDate:        req.DueDate,
			JournalEntryID: entry.ID,
		}
		if err := tx.Create(&ar).Error; err != nil {
			return fmt.Errorf("failed to create accounts receivable: %w", err)
		}
	}

	// Audit EVERY manual discount (who applied it, why, and how much).
	if manualDiscountAmount > 0 {
		s.emitManualDiscountAudit(tx, orderNumber, userID, req, manualDiscountAmount, subtotal)
	}
	s.emitDiscountAuditIfSignificant(tx, orderNumber, userID, subtotal, discountAmount, totalDiscountFromPromotions)

	return nil
}

// emitManualDiscountAudit records every employee-applied manual discount for
// accountability and review.
func (s *SalesService) emitManualDiscountAudit(
	tx *gorm.DB,
	orderNumber string,
	userID uint,
	req requests.CheckoutRequest,
	amount float64,
	subtotal float64,
) {
	pct := 0.0
	if subtotal > 0 {
		pct = (amount / subtotal) * 100.0
	}
	changes, err := json.Marshal(map[string]interface{}{
		"type":            req.ManualDiscountType,
		"value":           req.ManualDiscountValue,
		"amount":          round2(amount),
		"discount_pct":    round2(pct),
		"reason":          req.ManualDiscountReason,
		"applied_by":      userID,
		"order_subtotal":  round2(subtotal),
	})
	if err != nil {
		changes = []byte(`{}`)
	}
	entry := &domain.AuditLog{
		Action:    "MANUAL_DISCOUNT",
		Entity:    "Order",
		EntityID:  orderNumber,
		UserID:    userID,
		Changes:   string(changes),
		Timestamp: time.Now(),
	}
	_ = repository.WriteAudit(tx.Statement.Context, tx, entry)
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
