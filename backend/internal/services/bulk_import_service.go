package services

import (
	"encoding/csv"
	"fmt"
	"inventory/backend/internal/domain"
	"io"
	"os"
	"strconv"
)

type BulkImportResult struct {
	TotalRecords   int
	ValidRecords   int
	InvalidRecords int
	Errors         []error
	ValidProducts  []domain.Product
}

func ProcessBulkImport(filePath string) (*BulkImportResult, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Read header row
	_, err = reader.Read()
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
			result.Errors = append(result.Errors, fmt.Errorf("failed to read row: %w", err))
			continue
		}

		result.TotalRecords++

		product, validationErrors := validateAndParseProduct(row)
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

func validateAndParseProduct(row []string) (*domain.Product, []error) {
	var errors []error

	if len(row) != 12 {
		errors = append(errors, fmt.Errorf("invalid number of columns"))
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
		errors = append(errors, fmt.Errorf("SKU is required"))
	}
	if name == "" {
		errors = append(errors, fmt.Errorf("name is required"))
	}

	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		errors = append(errors, fmt.Errorf("invalid category ID: %w", err))
	}

	subCategoryID, err := strconv.ParseUint(subCategoryIDStr, 10, 32)
	if err != nil {
		errors = append(errors, fmt.Errorf("invalid sub-category ID: %w", err))
	}

	supplierID, err := strconv.ParseUint(supplierIDStr, 10, 32)
	if err != nil {
		errors = append(errors, fmt.Errorf("invalid supplier ID: %w", err))
	}

	purchasePrice, err := strconv.ParseFloat(purchasePriceStr, 64)
	if err != nil {
		errors = append(errors, fmt.Errorf("invalid purchase price: %w", err))
	}

	sellingPrice, err := strconv.ParseFloat(sellingPriceStr, 64)
	if err != nil {
		errors = append(errors, fmt.Errorf("invalid selling price: %w", err))
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
