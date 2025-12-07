package repository

import (
	"inventory/backend/internal/domain"
	"gorm.io/gorm"
)

type BarcodeRepository interface {
	GetProductBySKU(sku string) (*domain.Product, error)
	GetProductByID(id uint64) (*domain.Product, error)
	GetProductByBarcode(barcode string) (*domain.Product, error)
}

type barcodeRepository struct {
	db *gorm.DB
}

func NewBarcodeRepository(db *gorm.DB) BarcodeRepository {
	return &barcodeRepository{db: db}
}

func (r *barcodeRepository) GetProductBySKU(sku string) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.Where("sku = ?", sku).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *barcodeRepository) GetProductByID(id uint64) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *barcodeRepository) GetProductByBarcode(barcode string) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.Where("sku = ? OR barcode_upc = ?", barcode, barcode).First(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}
