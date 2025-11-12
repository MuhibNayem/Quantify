package services

import (
	"encoding/csv"
	"fmt"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"io"
	"strconv"
	"strings"
)

// BulkImportService handles the logic of processing a bulk import file.
type BulkImportService struct {
	categoryRepo *repository.CategoryRepository
	supplierRepo *repository.SupplierRepository
}

// NewBulkImportService creates a new BulkImportService.
func NewBulkImportService(categoryRepo *repository.CategoryRepository, supplierRepo *repository.SupplierRepository) *BulkImportService {
	return &BulkImportService{
		categoryRepo: categoryRepo,
		supplierRepo: supplierRepo,
	}
}

// NewEntities holds slices of new entities discovered during the import validation.
type NewEntities struct {
	Categories    map[string]bool `json:"categories"`
	SubCategories map[string]bool `json:"subCategories"`
	Suppliers     map[string]bool `json:"suppliers"`
}

// BulkImportResult holds the result of a bulk import validation.
type BulkImportResult struct {
	TotalRecords   int              `json:"totalRecords"`
	ValidRecords   int              `json:"validRecords"`
	InvalidRecords int              `json:"invalidRecords"`
	Errors         []string         `json:"errors"`
	ValidProducts  []domain.Product `json:"validProducts"`
	NewEntities    NewEntities      `json:"newEntities"`
}

// entityCache holds in-memory maps of existing entities to speed up validation.
type entityCache struct {
	categories    map[string]uint
	subCategories map[string]uint
	suppliers     map[string]uint
}

// ProcessBulkImport reads a CSV file and produces a validation result.
func (s *BulkImportService) ProcessBulkImport(file io.Reader) (*BulkImportResult, error) {
	reader := csv.NewReader(file)
	// Read header row
	_, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header row: %w", err)
	}

	result := &BulkImportResult{
		NewEntities: NewEntities{
			Categories:    make(map[string]bool),
			SubCategories: make(map[string]bool),
			Suppliers:     make(map[string]bool),
		},
	}

	// Pre-fetch existing entities into a cache
	cache, err := s.buildEntityCache()
	if err != nil {
		return nil, fmt.Errorf("failed to build entity cache: %w", err)
	}

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

		product, validationErrors := s.validateAndParseProduct(row, cache, &result.NewEntities)
		if len(validationErrors) > 0 {
			result.InvalidRecords++
			for _, e := range validationErrors {
				result.Errors = append(result.Errors, fmt.Sprintf("Row %d: %s", result.TotalRecords+1, e))
			}
		} else {
			result.ValidRecords++
			result.ValidProducts = append(result.ValidProducts, *product)
		}
	}

	return result, nil
}

// validateAndParseProduct validates a single CSV row and maps it to a product domain model.
func (s *BulkImportService) validateAndParseProduct(row []string, cache *entityCache, newEntities *NewEntities) (*domain.Product, []string) {
	var errors []string

	if len(row) != 11 {
		return nil, []string{"invalid number of columns"}
	}

	// Assign columns to variables
	sku, name, description := row[0], row[1], row[2]
	categoryName, subCategoryName, supplierName := row[3], row[4], row[5]
	brand := row[6]
	purchasePriceStr, sellingPriceStr := row[7], row[8]
	locationName := row[9]
	status := row[10]

	// Basic validations
	if sku == "" {
		errors = append(errors, "SKU is required")
	}
	if name == "" {
		errors = append(errors, "Name is required")
	}

	// --- Category, SubCategory, and Supplier validation ---
	var categoryID, subCategoryID, supplierID uint
	var ok bool

	if categoryName != "" {
		if categoryID, ok = cache.categories[strings.ToLower(categoryName)]; !ok {
			newEntities.Categories[categoryName] = true // Mark for creation
		}
	}

	if subCategoryName != "" {
		if subCategoryID, ok = cache.subCategories[strings.ToLower(subCategoryName)]; !ok {
			if categoryName == "" {
				errors = append(errors, "SubCategory cannot be created without a parent Category")
			} else {
				newEntities.SubCategories[subCategoryName] = true // Mark for creation
			}
		}
	}

	if supplierName != "" {
		if supplierID, ok = cache.suppliers[strings.ToLower(supplierName)]; !ok {
			newEntities.Suppliers[supplierName] = true // Mark for creation
		}
	}

	if strings.TrimSpace(locationName) == "" {
		errors = append(errors, "LocationName is required")
	}

	// --- Price validation ---
	purchasePrice, err := strconv.ParseFloat(purchasePriceStr, 64)
	if err != nil && purchasePriceStr != "" {
		errors = append(errors, fmt.Sprintf("invalid purchase price: %v", err))
	}

	sellingPrice, err := strconv.ParseFloat(sellingPriceStr, 64)
	if err != nil && sellingPriceStr != "" {
		errors = append(errors, fmt.Sprintf("invalid selling price: %v", err))
	}

	if len(errors) > 0 {
		return nil, errors
	}

	// Return a temporary product object. IDs will be filled in after creation.
	return &domain.Product{
		SKU:           sku,
		Name:          name,
		Description:   description,
		CategoryID:    categoryID,    // Will be 0 for new categories
		SubCategoryID: subCategoryID, // Will be 0 for new sub-categories
		SupplierID:    supplierID,    // Will be 0 for new suppliers
		Brand:         brand,
		PurchasePrice: purchasePrice,
		SellingPrice:  sellingPrice,
		Status:        status,
		Category:      domain.Category{Name: categoryName},
		Supplier:      domain.Supplier{Name: supplierName},
		SubCategory:   domain.SubCategory{Name: subCategoryName},
		Location:      domain.Location{Name: locationName},
	}, nil
}

// buildEntityCache fetches existing entities from the database to speed up validation.
func (s *BulkImportService) buildEntityCache() (*entityCache, error) {
	cache := &entityCache{
		categories:    make(map[string]uint),
		subCategories: make(map[string]uint),
		suppliers:     make(map[string]uint),
	}

	// Cache Categories
	cats, err := s.categoryRepo.GetAll()
	if err != nil {
		return nil, err
	}
	for _, cat := range cats {
		cache.categories[strings.ToLower(cat.Name)] = cat.ID
	}

	// Cache SubCategories
	subCats, err := s.categoryRepo.GetAllSubCategories()
	if err != nil {
		return nil, err
	}
	for _, subCat := range subCats {
		cache.subCategories[strings.ToLower(subCat.Name)] = subCat.ID
	}

	// Cache Suppliers
	sups, err := s.supplierRepo.GetAll()
	if err != nil {
		return nil, err
	}
	for _, sup := range sups {
		cache.suppliers[strings.ToLower(sup.Name)] = sup.ID
	}

	return cache, nil
}
