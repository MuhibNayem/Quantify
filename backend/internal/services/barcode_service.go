package services

import (
	"bytes"
	"image/png"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
)

type BarcodeService interface {
	GenerateBarcode(sku string, productID uint64) (*bytes.Buffer, error)
	LookupProductByBarcode(barcodeValue string) (*domain.Product, error)
}

type barcodeService struct {
	repo repository.BarcodeRepository
}

func NewBarcodeService(repo repository.BarcodeRepository) BarcodeService {
	return &barcodeService{repo: repo}
}

func (s *barcodeService) GenerateBarcode(sku string, productID uint64) (*bytes.Buffer, error) {
	var product *domain.Product
	var err error

	if sku != "" {
		product, err = s.repo.GetProductBySKU(sku)
	} else if productID != 0 {
		product, err = s.repo.GetProductByID(productID)
	}
	if err != nil {
		return nil, err
	}

	content := product.SKU
	if product.BarcodeUPC != "" {
		content = product.BarcodeUPC
	}

	// Create the barcode
	var bcode barcode.Barcode
	bcode, err = code128.Encode(content)
	if err != nil {
		return nil, err
	}

	// Scale the barcode
	bcode, err = barcode.Scale(bcode, 200, 50) // Width, Height
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, bcode); err != nil {
		return nil, err
	}

	return &buf, nil
}

func (s *barcodeService) LookupProductByBarcode(barcodeValue string) (*domain.Product, error) {
	return s.repo.GetProductByBarcode(barcodeValue)
}
