package repository

import (
	"inventory/backend/internal/domain"

	"gorm.io/gorm"
)

type ReplenishmentRepository interface {
	GetProductStock(productID uint) (int, error)
	GetAllProductAlertSettings() ([]domain.ProductAlertSettings, error)
	GetPendingSuggestion(productID uint) (*domain.ReorderSuggestion, error)
	GetPendingPO(productID uint) (*domain.PurchaseOrder, error)
	CreateReorderSuggestion(suggestion *domain.ReorderSuggestion) error
	GetSupplierForProduct(productID uint) (*domain.Supplier, error)
	GetStockLevels(productIDs []uint) (map[uint]int, error)
	GetPendingSuggestionsMap(productIDs []uint) (map[uint]bool, error)
	GetPendingPOsMap(productIDs []uint) (map[uint]bool, error)
	ListReorderSuggestions(status string, supplierID string) ([]domain.ReorderSuggestion, error)
}

type replenishmentRepository struct {
	db *gorm.DB
}

func NewReplenishmentRepository(db *gorm.DB) ReplenishmentRepository {
	return &replenishmentRepository{db: db}
}

func (r *replenishmentRepository) ListReorderSuggestions(status string, supplierID string) ([]domain.ReorderSuggestion, error) {
	var suggestions []domain.ReorderSuggestion
	db := r.db.Preload("Product").Preload("Supplier")

	if status != "" {
		db = db.Where("status = ?", status)
	}
	if supplierID != "" {
		db = db.Where("supplier_id = ?", supplierID)
	}

	err := db.Find(&suggestions).Error
	return suggestions, err
}

func (r *replenishmentRepository) GetProductStock(productID uint) (int, error) {
	var totalStock int64
	err := r.db.Model(&domain.Batch{}).
		Where("product_id = ?", productID).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&totalStock).Error
	return int(totalStock), err
}

func (r *replenishmentRepository) GetAllProductAlertSettings() ([]domain.ProductAlertSettings, error) {
	var settings []domain.ProductAlertSettings
	// Preload Product to get SupplierID later if needed, or just ProductID
	err := r.db.Preload("Product").Find(&settings).Error
	return settings, err
}

func (r *replenishmentRepository) GetPendingSuggestion(productID uint) (*domain.ReorderSuggestion, error) {
	var suggestion domain.ReorderSuggestion
	err := r.db.Where("product_id = ? AND status = ?", productID, "PENDING").First(&suggestion).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &suggestion, nil
}

func (r *replenishmentRepository) GetPendingPO(productID uint) (*domain.PurchaseOrder, error) {
	// Check for POs that are active (not received or cancelled) containing this product
	// This requires joining PurchaseOrderItems
	var po domain.PurchaseOrder
	err := r.db.Joins("JOIN purchase_order_items ON purchase_order_items.purchase_order_id = purchase_orders.id").
		Where("purchase_order_items.product_id = ?", productID).
		Where("purchase_orders.status IN ?", []string{"DRAFT", "APPROVED", "SENT", "PARTIALLY_RECEIVED"}).
		First(&po).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &po, nil
}

func (r *replenishmentRepository) CreateReorderSuggestion(suggestion *domain.ReorderSuggestion) error {
	return r.db.Create(suggestion).Error
}

func (r *replenishmentRepository) GetSupplierForProduct(productID uint) (*domain.Supplier, error) {
	var product domain.Product
	if err := r.db.Preload("Supplier").First(&product, productID).Error; err != nil {
		return nil, err
	}
	return &product.Supplier, nil
}

func (r *replenishmentRepository) GetStockLevels(productIDs []uint) (map[uint]int, error) {
	type Result struct {
		ProductID uint
		Total     int
	}
	var results []Result
	err := r.db.Model(&domain.Batch{}).
		Where("product_id IN ?", productIDs).
		Select("product_id, COALESCE(SUM(quantity), 0) as total").
		Group("product_id").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	stockMap := make(map[uint]int)
	for _, res := range results {
		stockMap[res.ProductID] = res.Total
	}
	return stockMap, nil
}

func (r *replenishmentRepository) GetPendingSuggestionsMap(productIDs []uint) (map[uint]bool, error) {
	var suggestions []domain.ReorderSuggestion
	err := r.db.Where("product_id IN ? AND status = ?", productIDs, "PENDING").
		Select("product_id").
		Find(&suggestions).Error

	if err != nil {
		return nil, err
	}

	pendingMap := make(map[uint]bool)
	for _, s := range suggestions {
		pendingMap[s.ProductID] = true
	}
	return pendingMap, nil
}

func (r *replenishmentRepository) GetPendingPOsMap(productIDs []uint) (map[uint]bool, error) {
	var poItems []domain.PurchaseOrderItem
	err := r.db.Joins("JOIN purchase_orders ON purchase_orders.id = purchase_order_items.purchase_order_id").
		Where("purchase_order_items.product_id IN ?", productIDs).
		Where("purchase_orders.status IN ?", []string{"DRAFT", "APPROVED", "SENT", "PARTIALLY_RECEIVED"}).
		Select("purchase_order_items.product_id").
		Find(&poItems).Error

	if err != nil {
		return nil, err
	}

	pendingMap := make(map[uint]bool)
	for _, item := range poItems {
		pendingMap[item.ProductID] = true
	}
	return pendingMap, nil
}
