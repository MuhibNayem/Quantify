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
	GenerateBarcode(sku string) (*bytes.Buffer, error)
	LookupProductByBarcode(barcodeValue string) (*domain.Product, error)
	ResolveBarcode(barcodeValue string) (*domain.Product, *domain.ProductUnit, error)
}

type barcodeService struct {
	repo repository.BarcodeRepository
}

func NewBarcodeService(repo repository.BarcodeRepository) BarcodeService {
	return &barcodeService{repo: repo}
}

func (s *barcodeService) GenerateBarcode(sku string) (*bytes.Buffer, error) {
	product, err := s.repo.GetProductBySKU(sku)
	if err != nil {
		return nil, err
	}

	content := product.SKU
	if product.BarcodeUPC != "" {
		content = product.BarcodeUPC
	}

	var bcode barcode.Barcode
	bcode, err = code128.Encode(content)
	if err != nil {
		return nil, err
	}
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
	product, _, err := s.repo.ResolveBarcode(barcodeValue)
	return product, err
}

func (s *barcodeService) ResolveBarcode(barcodeValue string) (*domain.Product, *domain.ProductUnit, error) {
	return s.repo.ResolveBarcode(barcodeValue)
}
