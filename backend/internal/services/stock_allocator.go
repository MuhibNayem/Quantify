package services

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/requests"
)

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

func (s *SalesService) allocateStock(
	tx *gorm.DB,
	product domain.Product,
	batches []*domain.Batch,
	locationID uint,
	quantity int,
	customerID *uint,
	userID uint,
) (int, error) {
	availableStock := 0
	for _, b := range batches {
		availableStock += b.Quantity
	}

	if availableStock < quantity {
		locLabel := "any location"
		if locationID != 0 {
			locLabel = fmt.Sprintf("location %d", locationID)
		}
		return 0, fmt.Errorf("insufficient stock for product '%s' at %s (Available: %d, Requested: %d)", product.Name, locLabel, availableStock, quantity)
	}

	qtyToReduce := quantity
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
			return 0, fmt.Errorf("failed to update batch %s", batch.BatchNumber)
		}
	}

	adjustment := domain.StockAdjustment{
		ProductID:        product.ID,
		LocationID:       locationID,
		Type:             "STOCK_OUT",
		Quantity:         quantity,
		ReasonCode:       "SALE",
		Notes:            fmt.Sprintf("Sale to customer ID: %d", customerID),
		AdjustedBy:       userID,
		AdjustedAt:       time.Now(),
		PreviousQuantity: availableStock,
		NewQuantity:      availableStock - quantity,
	}
	if err := tx.Create(&adjustment).Error; err != nil {
		return 0, fmt.Errorf("failed to create stock adjustment: %w", err)
	}

	return availableStock, nil
}
