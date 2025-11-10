package repository

import (
	"inventory/backend/internal/domain"

	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) GetProductBySKU(sku string) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.Where("sku = ?", sku).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *ProductRepository) GetProductByBarcode(barcode string) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.Where("barcode_upc = ?", barcode).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}
