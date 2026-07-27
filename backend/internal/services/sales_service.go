package services

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
)

// discountAuditThresholdPct is the minimum combined discount share (promotions
// + loyalty points) of the pre-discount subtotal that triggers an audit row.
// Transactions at or above this threshold are written to audit_log so that
// "manager-priced" sales leave a compliance trail.
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
		// 1. Fetch Data
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

		// 2. Process Items (Stock & Discounts)
		subtotal, totalDiscountFromPromotions, orderItems, err := s.processItems(tx, req, products, batchesByProduct, activePromotions, userID)
		if err != nil {
			return err
		}

		// 3. Apply Tax (returns post-tax total and the tax portion separately so
		//    Order.Subtotal and Order.TaxAmount can be persisted alongside TotalAmount)
		postTaxTotal, taxAmount, err := s.applyTax(subtotal)
		if err != nil {
			return err
		}

		// 4. Loyalty Points
		discountAmount, pointsRedeemed, err := s.handleLoyalty(tx, req.CustomerID, req.PointsToRedeem, postTaxTotal)
		if err != nil {
			return err
		}

		// 5. Create Order & Transaction
		return s.createOrderAndTransaction(tx, req, userID, subtotal, taxAmount, discountAmount, totalDiscountFromPromotions, pointsRedeemed, orderItems)
	})
}

// resolveDeductionLocation decides which location to deduct stock from.
// When a register overrides the location, use it; otherwise use the product's default location.
func resolveDeductionLocation(product domain.Product, registerLocationID *uint) uint {
	if registerLocationID != nil && *registerLocationID != 0 {
		return *registerLocationID
	}
	return product.LocationID
}

func (s *SalesService) fetchProducts(tx *gorm.DB, items []requests.CheckoutItem) (map[uint]domain.Product, error) {
	productIDs := make([]uint, len(items))
	for i, item := range items {
		productIDs[i] = item.ProductID
	}

	var products []domain.Product
	if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("id IN ?", productIDs).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	if len(products) != len(items) {
		return nil, fmt.Errorf("some products not found")
	}

	productMap := make(map[uint]domain.Product)
	for _, p := range products {
		productMap[p.ID] = p
	}
	return productMap, nil
}

func (s *SalesService) fetchBatches(tx *gorm.DB, items []requests.CheckoutItem) (map[uint][]*domain.Batch, error) {
	productIDs := make([]uint, len(items))
	for i, item := range items {
		productIDs[i] = item.ProductID
	}

	var allBatches []domain.Batch
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("product_id IN ? AND quantity > 0", productIDs).
		Order("expiry_date ASC NULLS LAST, created_at ASC").
		Find(&allBatches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch batches: %w", err)
	}

	batchesByProduct := make(map[uint][]*domain.Batch)
	for i := range allBatches {
		batchesByProduct[allBatches[i].ProductID] = append(batchesByProduct[allBatches[i].ProductID], &allBatches[i])
	}
	return batchesByProduct, nil
}

// fetchBatchesForLocation returns available batches for a single product at a specific location,
// locked for update so concurrent checkouts cannot oversell.
func (s *SalesService) fetchBatchesForLocation(tx *gorm.DB, productID, locationID uint) ([]*domain.Batch, error) {
	var batches []domain.Batch
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("product_id = ? AND quantity > 0", productID).
		Where("location_id = ?", locationID).
		Order("expiry_date ASC NULLS LAST, created_at ASC").
		Find(&batches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch batches: %w", err)
	}

	out := make([]*domain.Batch, len(batches))
	for i := range batches {
		out[i] = &batches[i]
	}
	return out, nil
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
	var stockAdjustments []domain.StockAdjustment

	for _, item := range req.Items {
		product, ok := productMap[item.ProductID]
		if !ok {
			return 0, 0, nil, fmt.Errorf("product %d not found", item.ProductID)
		}

		// Decide which location to deduct stock from for this line item.
		deductionLocationID := resolveDeductionLocation(product, req.RegisterLocationID)

		// Prefer batches at the resolved location; fall back to the cached set if nothing is found locally.
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

		availableStock := 0
		for _, b := range batches {
			availableStock += b.Quantity
		}

		if availableStock < item.Quantity {
			locLabel := "any location"
			if deductionLocationID != 0 {
				locLabel = fmt.Sprintf("location %d", deductionLocationID)
			}
			return 0, 0, nil, fmt.Errorf("insufficient stock for product '%s' at %s (Available: %d, Requested: %d)", product.Name, locLabel, availableStock, item.Quantity)
		}

		qtyToReduce := item.Quantity
		for _, batch := range batches {
			if qtyToReduce <= 0 {
				break
			}
			if batch.Quantity >= qtyToReduce {
				batch.Quantity -= qtyToReduce
				qtyToReduce = 0
			} else {
				qtyToReduce -= batch.Quantity
				batch.Quantity = 0
			}
			if err := tx.Save(batch).Error; err != nil {
				return 0, 0, nil, fmt.Errorf("failed to update batch %s", batch.BatchNumber)
			}
		}

		stockAdjustments = append(stockAdjustments, domain.StockAdjustment{
			ProductID:        product.ID,
			LocationID:       deductionLocationID,
			Type:             "STOCK_OUT",
			Quantity:         item.Quantity,
			ReasonCode:       "SALE",
			Notes:            fmt.Sprintf("Sale to customer ID: %d", req.CustomerID),
			AdjustedBy:       userID,
			AdjustedAt:       time.Now(),
			PreviousQuantity: availableStock,
			NewQuantity:      availableStock - item.Quantity,
		})

		// Calculate Discount
		unitPrice, discount := s.calculateItemDiscount(product, item.Quantity, activePromotions)
		totalDiscountFromPromotions += discount
		totalAmount += unitPrice * float64(item.Quantity)

		orderItems = append(orderItems, domain.OrderItem{
			ProductID:  item.ProductID,
			Quantity:   item.Quantity,
			UnitPrice:  unitPrice,
			TotalPrice: unitPrice * float64(item.Quantity),
		})
	}

	// Bulk Create Stock Adjustments
	if len(stockAdjustments) > 0 {
		if err := tx.Create(&stockAdjustments).Error; err != nil {
			return 0, 0, nil, fmt.Errorf("failed to create stock adjustments: %w", err)
		}
	}

	return totalAmount, totalDiscountFromPromotions, orderItems, nil
}

func (s *SalesService) calculateItemDiscount(product domain.Product, quantity int, activePromotions []domain.Promotion) (float64, float64) {
	bestPromo := ResolvePromotion(activePromotions, product.ID, product.SubCategoryID, product.CategoryID)

	unitPrice := product.SellingPrice
	discountTotal := 0.0
	if bestPromo != nil {
		if bestPromo.DiscountType == "PERCENTAGE" {
			discount := unitPrice * (bestPromo.DiscountValue / 100.0)
			unitPrice -= discount
		} else if bestPromo.DiscountType == "FIXED_AMOUNT" {
			unitPrice -= bestPromo.DiscountValue
		}
		if unitPrice < 0 {
			unitPrice = 0
		}
		discountTotal = (product.SellingPrice - unitPrice) * float64(quantity)
	}
	return unitPrice, discountTotal
}

// applyTax returns the post-tax total and the tax amount computed from the
// configured tax_rate_percentage. The pre-tax amount is passed in as `amount`.
// Per P3.T7 the order row stores both subtotal and tax amount so downstream
// reports do not need to derive tax from total / (1 + rate).
func (s *SalesService) applyTax(amount float64) (postTaxTotal, taxAmount float64, err error) {
	taxRate := 0.0
	if val, err := s.Settings.GetSetting("tax_rate_percentage"); err == nil {
		if v, err := strconv.ParseFloat(val, 64); err == nil {
			taxRate = v / 100.0
		}
	}
	taxAmount = amount * taxRate
	postTaxTotal = amount + taxAmount
	return postTaxTotal, taxAmount, nil
}

func (s *SalesService) handleLoyalty(tx *gorm.DB, customerID *uint, pointsToRedeem int, totalAmount float64) (float64, int, error) {
	if customerID == nil || *customerID == 0 {
		if pointsToRedeem > 0 {
			return 0, 0, fmt.Errorf("cannot redeem points without a customer selected")
		}
		return 0, 0, nil
	}

	var loyalty domain.LoyaltyAccount
	result := tx.Where("user_id = ?", *customerID).First(&loyalty)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			loyalty = domain.LoyaltyAccount{
				UserID: *customerID,
				Points: 0,
				Tier:   "Bronze",
			}
			if err := tx.Create(&loyalty).Error; err != nil {
				return 0, 0, fmt.Errorf("failed to create loyalty account: %w", err)
			}
		} else {
			return 0, 0, fmt.Errorf("failed to fetch loyalty account: %w", result.Error)
		}
	}

	discountAmount := 0.0
	pointsRedeemed := 0

	// Redemption
	if pointsToRedeem > 0 {
		if loyalty.Points < pointsToRedeem {
			return 0, 0, fmt.Errorf("insufficient loyalty points (Available: %d, Requested: %d)", loyalty.Points, pointsToRedeem)
		}

		redemptionRate := 0.01
		if val, err := s.Settings.GetSetting("loyalty_points_redemption_rate"); err == nil {
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				redemptionRate = v
			}
		}

		discountAmount = float64(pointsToRedeem) * redemptionRate
		if discountAmount > totalAmount {
			return 0, 0, fmt.Errorf("loyalty discount amount (%.2f) exceeds order total (%.2f)", discountAmount, totalAmount)
		}

		loyalty.Points -= pointsToRedeem
		pointsRedeemed = pointsToRedeem
	}

	// Earning
	earningRate := 1.0
	if val, err := s.Settings.GetSetting("loyalty_points_earning_rate"); err == nil {
		if v, err := strconv.ParseFloat(val, 64); err == nil {
			earningRate = v
		}
	}

	netAmount := totalAmount - discountAmount
	if netAmount < 0 {
		netAmount = 0
	}

	pointsEarned := int(netAmount * earningRate)
	loyalty.Points += pointsEarned

	// Update Tier
	s.updateLoyaltyTier(&loyalty)

	if err := tx.Save(&loyalty).Error; err != nil {
		return 0, 0, fmt.Errorf("failed to update loyalty points: %w", err)
	}

	return discountAmount, pointsRedeemed, nil
}

func (s *SalesService) updateLoyaltyTier(loyalty *domain.LoyaltyAccount) {
	silverThreshold := 500
	goldThreshold := 2500
	platinumThreshold := 10000

	if val, err := s.Settings.GetSetting("loyalty_tier_silver"); err == nil {
		if v, err := strconv.Atoi(val); err == nil {
			silverThreshold = v
		}
	}
	if val, err := s.Settings.GetSetting("loyalty_tier_gold"); err == nil {
		if v, err := strconv.Atoi(val); err == nil {
			goldThreshold = v
		}
	}
	if val, err := s.Settings.GetSetting("loyalty_tier_platinum"); err == nil {
		if v, err := strconv.Atoi(val); err == nil {
			platinumThreshold = v
		}
	}

	if loyalty.Points >= platinumThreshold {
		loyalty.Tier = "Platinum"
	} else if loyalty.Points >= goldThreshold {
		loyalty.Tier = "Gold"
	} else if loyalty.Points >= silverThreshold {
		loyalty.Tier = "Silver"
	}
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

// emitDiscountAuditIfSignificant writes an audit_log row when the combined
// promotion + loyalty discount is at least discountAuditThresholdPct of the
// pre-discount subtotal. The call is best-effort: any failure is logged
// inside WriteAudit and the function returns without surfacing an error, so
// an audit-write failure never aborts a successful checkout.
//
// `subtotalAfterPromo` is the pre-tax, post-promo-discount amount. The
// pre-discount subtotal is reconstructed by adding `promotionDiscount` back.
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
		"subtotal":              round2(preDiscountSubtotal),
		"total_discount":        round2(totalDiscount),
		"discount_pct":          round2(discountPct),
		"promotion_discount":    round2(promotionDiscount),
		"loyalty_discount":      round2(loyaltyDiscount),
	})
	if err != nil {
		// Should not happen for plain float64; fall through with empty Changes.
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

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
