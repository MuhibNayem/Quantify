package services

import (
	"encoding/csv"
	"fmt"
	"inventory/backend/internal/domain"
	"io"
	"strconv"
)

type BulkImportService struct{}

func NewBulkImportService() *BulkImportService {
	return &BulkImportService{}
}

type BulkImportResult struct {
	TotalRecords   int              `json:"totalRecords"`
	ValidRecords   int              `json:"validRecords"`
	InvalidRecords int              `json:"invalidRecords"`
	Errors         []string         `json:"errors"`
	ValidProducts  []domain.Product `json:"validProducts"`
}

func (s *BulkImportService) ProcessBulkImport(file io.Reader) (*BulkImportResult, error) {
	reader := csv.NewReader(file)
	// Read header row
	_, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header row: %w", err)
	}

	result := &BulkImportResult{}

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.InvalidRecords++
			result.Errors = append(result.Errors, fmt.Sprintf("failed to read row: %v", err))
			continue
		}

		result.TotalRecords++

		product, validationErrors := s.validateAndParseProduct(row)
		if len(validationErrors) > 0 {
			result.InvalidRecords++
			result.Errors = append(result.Errors, validationErrors...)
		} else {
			result.ValidRecords++
			result.ValidProducts = append(result.ValidProducts, *product)
		}
	}

	return result, nil
}

func (s *BulkImportService) validateAndParseProduct(row []string) (*domain.Product, []string) {
	var errors []string

	if len(row) != 12 {
		errors = append(errors, "invalid number of columns")
		return nil, errors
	}

	sku := row[0]
	name := row[1]
	description := row[2]
	categoryIDStr := row[3]
	subCategoryIDStr := row[4]
	supplierIDStr := row[5]
	brand := row[6]
	purchasePriceStr := row[7]
	sellingPriceStr := row[8]
	barcodeUPC := row[9]
	imageURLs := row[10]
	status := row[11]

	if sku == "" {
		errors = append(errors, "SKU is required")
	}
	if name == "" {
		errors = append(errors, "name is required")
	}

	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		errors = append(errors, fmt.Sprintf("invalid category ID: %v", err))
	}

	subCategoryID, err := strconv.ParseUint(subCategoryIDStr, 10, 32)
	if err != nil {
		errors = append(errors, fmt.Sprintf("invalid sub-category ID: %v", err))
	}

	supplierID, err := strconv.ParseUint(supplierIDStr, 10, 32)
	if err != nil {
		errors = append(errors, fmt.Sprintf("invalid supplier ID: %v", err))
	}

	purchasePrice, err := strconv.ParseFloat(purchasePriceStr, 64)
	if err != nil {
		errors = append(errors, fmt.Sprintf("invalid purchase price: %v", err))
	}

	sellingPrice, err := strconv.ParseFloat(sellingPriceStr, 64)
	if err != nil {
		errors = append(errors, fmt.Sprintf("invalid selling price: %v", err))
	}

	if len(errors) > 0 {
		return nil, errors
	}

	return &domain.Product{
		SKU:           sku,
		Name:          name,
		Description:   description,
		CategoryID:    uint(categoryID),
		SubCategoryID: uint(subCategoryID),
		SupplierID:    uint(supplierID),
		Brand:         brand,
		PurchasePrice: purchasePrice,
		SellingPrice:  sellingPrice,
		BarcodeUPC:    barcodeUPC,
		ImageURLs:     imageURLs,
		Status:        status,
	}, nil
}
