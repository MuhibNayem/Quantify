package repository

import (
	"gorm.io/gorm"
	"inventory/backend/internal/domain"
)

type BarcodeRepository interface {
	GetProductBySKU(sku string) (*domain.Product, error)
	GetProductByBarcode(barcode string) (*domain.Product, error)
	// ResolveBarcode returns the scanned product and, when the barcode belongs
	// to a selling unit, the matching ProductUnit (nil for product-level codes).
	ResolveBarcode(barcode string) (*domain.Product, *domain.ProductUnit, error)
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

func (r *barcodeRepository) GetProductByBarcode(barcode string) (*domain.Product, error) {
	product, _, err := r.ResolveBarcode(barcode)
	return product, err
}

func (r *barcodeRepository) ResolveBarcode(barcode string) (*domain.Product, *domain.ProductUnit, error) {
	// Product-level barcode wins over SKU, then SKU, then unit barcode.
	var product domain.Product
	err := r.db.Where("barcode_upc = ?", barcode).First(&product).Error
	if err == gorm.ErrRecordNotFound {
		err = r.db.Where("sku = ?", barcode).First(&product).Error
	}
	if err == nil {
		return &product, nil, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, nil, err
	}

	// Otherwise look for an active per-selling-unit barcode (a coil and a foot
	// of the same product scan differently).
	var unit domain.ProductUnit
	if err := r.db.Preload("Unit").Preload("Product").
		Where("barcode = ? AND is_active = ?", barcode, true).
		First(&unit).Error; err != nil {
		return nil, nil, err
	}

	product = unit.Product
	return &product, &unit, nil
}
