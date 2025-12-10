package services

import (
	"fmt"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/requests"
	"strconv"
	"time"

	"gorm.io/gorm"
)

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
		totalAmount, totalDiscountFromPromotions, orderItems, err := s.processItems(tx, req, products, batchesByProduct, activePromotions, userID)
		if err != nil {
			return err
		}

		// 3. Apply Tax
		totalAmount, err = s.applyTax(totalAmount)
		if err != nil {
			return err
		}

		// 4. Loyalty Points
		discountAmount, pointsRedeemed, err := s.handleLoyalty(tx, req.CustomerID, req.PointsToRedeem, totalAmount)
		if err != nil {
			return err
		}

		// 5. Create Order & Transaction
		return s.createOrderAndTransaction(tx, req, userID, totalAmount, discountAmount, totalDiscountFromPromotions, pointsRedeemed, orderItems)
	})
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
	if err := tx.Where("product_id IN ? AND quantity > 0", productIDs).Order("expiry_date asc, created_at asc").Find(&allBatches).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch batches: %w", err)
	}

	batchesByProduct := make(map[uint][]*domain.Batch)
	for i := range allBatches {
		batchesByProduct[allBatches[i].ProductID] = append(batchesByProduct[allBatches[i].ProductID], &allBatches[i])
	}
	return batchesByProduct, nil
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

		// Stock Deduction
		batches := batchesByProduct[item.ProductID]
		availableStock := 0
		for _, b := range batches {
			availableStock += b.Quantity
		}

		if availableStock < item.Quantity {
			return 0, 0, nil, fmt.Errorf("insufficient stock for product '%s' (Available: %d, Requested: %d)", product.Name, availableStock, item.Quantity)
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
			LocationID:       product.LocationID,
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
	var bestPromo *domain.Promotion
	bestScore := -1

	for i := range activePromotions {
		p := &activePromotions[i]
		score := -1

		if p.ProductID != nil && *p.ProductID == product.ID {
			score = 2
		} else if p.SubCategoryID != nil && product.SubCategoryID > 0 && *p.SubCategoryID == product.SubCategoryID {
			score = 1
		} else if p.CategoryID != nil && product.CategoryID > 0 && *p.CategoryID == product.CategoryID {
			score = 0
		}

		if score > -1 {
			if score > bestScore {
				bestScore = score
				bestPromo = p
			} else if score == bestScore {
				if bestPromo != nil && p.Priority > bestPromo.Priority {
					bestPromo = p
				}
			}
		}
	}

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

func (s *SalesService) applyTax(amount float64) (float64, error) {
	taxRate := 0.0
	if val, err := s.Settings.GetSetting("tax_rate_percentage"); err == nil {
		if v, err := strconv.ParseFloat(val, 64); err == nil {
			taxRate = v / 100.0
		}
	}
	return amount + (amount * taxRate), nil
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
	totalAmount float64,
	discountAmount float64,
	totalDiscountFromPromotions float64,
	pointsRedeemed int,
	orderItems []domain.OrderItem,
) error {
	orderNumber := fmt.Sprintf("ORD-%d-%d", time.Now().Unix(), userID)
	order := domain.Order{
		OrderNumber:    orderNumber,
		UserID:         userID,
		CustomerID:     req.CustomerID,
		TotalAmount:    totalAmount - discountAmount,
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
		Amount:               int64((totalAmount - discountAmount) * 100),
		Currency:             "USD",
		PaymentMethod:        req.PaymentMethod,
		Status:               "COMPLETED",
		GatewayTransactionID: fmt.Sprintf("GW-%d", time.Now().UnixNano()),
	}

	if err := tx.Create(&saletransaction).Error; err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	return nil
}
