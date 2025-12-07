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
	locationRepo *repository.LocationRepository
}

// NewBulkImportService creates a new BulkImportService.
func NewBulkImportService(categoryRepo *repository.CategoryRepository, supplierRepo *repository.SupplierRepository, locationRepo *repository.LocationRepository) *BulkImportService {
	return &BulkImportService{
		categoryRepo: categoryRepo,
		supplierRepo: supplierRepo,
		locationRepo: locationRepo,
	}
}

// NewEntities holds slices of new entities discovered during the import validation.
type NewEntities struct {
	Categories    map[string]bool `json:"categories"`
	SubCategories map[string]bool `json:"subCategories"`
	Suppliers     map[string]bool `json:"suppliers"`
	Locations     map[string]bool `json:"locations"`
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
	locations     map[string]uint
}

// ProcessBulkImport reads a CSV file and produces a validation result.
func (s *BulkImportService) ProcessBulkImport(file io.ReadSeeker) (*BulkImportResult, error) {
	// Pass 1: Collect unique names to build optimized cache
	cache, err := s.buildOptimizedCache(file)
	if err != nil {
		return nil, fmt.Errorf("failed to build optimized cache: %w", err)
	}

	// Reset file pointer for Pass 2
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to reset file pointer: %w", err)
	}

	reader := csv.NewReader(file)
	// Read header row
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header row: %w", err)
	}

	result := &BulkImportResult{
		NewEntities: NewEntities{
			Categories:    make(map[string]bool),
			SubCategories: make(map[string]bool),
			Suppliers:     make(map[string]bool),
			Locations:     make(map[string]bool),
		},
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

// buildOptimizedCache scans the file to find referenced entities and fetches only them.
func (s *BulkImportService) buildOptimizedCache(file io.ReadSeeker) (*entityCache, error) {
	reader := csv.NewReader(file)
	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, err // Empty file or error
	}

	uniqueCategories := make(map[string]bool)
	uniqueSubCategories := make(map[string]bool)
	uniqueSuppliers := make(map[string]bool)
	uniqueLocations := make(map[string]bool)

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // Skip bad rows during cache build
		}
		if len(row) >= 6 {
			if cat := strings.TrimSpace(row[3]); cat != "" {
				uniqueCategories[strings.ToLower(cat)] = true
			}
			if sub := strings.TrimSpace(row[4]); sub != "" {
				uniqueSubCategories[strings.ToLower(sub)] = true
			}
			if sup := strings.TrimSpace(row[5]); sup != "" {
				uniqueSuppliers[strings.ToLower(sup)] = true
			}
			if loc := strings.TrimSpace(row[9]); loc != "" {
				uniqueLocations[strings.ToLower(loc)] = true
			}
		}
	}

	cache := &entityCache{
		categories:    make(map[string]uint),
		subCategories: make(map[string]uint),
		suppliers:     make(map[string]uint),
		locations:     make(map[string]uint),
	}

	// Fetch only referenced Categories
	if len(uniqueCategories) > 0 {
		var cats []domain.Category
		names := make([]string, 0, len(uniqueCategories))
		for name := range uniqueCategories {
			names = append(names, name)
		}
		// Note: This assumes case-insensitive collation or we need to handle case in DB.
		// For simplicity, we fetch by name IN (...) and map them.
		// In a real scenario, we might need LOWER(name) IN (...)
		// Here we assume the DB handles it or we fetch and match.
		// GORM doesn't support map keys in Where easily without raw SQL for LOWER.
		// We'll fetch all matches.
		if err := s.categoryRepo.GetByNames(names, &cats); err == nil {
			for _, c := range cats {
				cache.categories[strings.ToLower(c.Name)] = c.ID
			}
		}
	}

	// Fetch only referenced SubCategories
	if len(uniqueSubCategories) > 0 {
		var subCats []domain.SubCategory
		names := make([]string, 0, len(uniqueSubCategories))
		for name := range uniqueSubCategories {
			names = append(names, name)
		}
		if err := s.categoryRepo.GetSubCategoriesByNames(names, &subCats); err == nil {
			for _, sc := range subCats {
				cache.subCategories[strings.ToLower(sc.Name)] = sc.ID
			}
		}
	}

	// Fetch only referenced Suppliers
	if len(uniqueSuppliers) > 0 {
		var sups []domain.Supplier
		names := make([]string, 0, len(uniqueSuppliers))
		for name := range uniqueSuppliers {
			names = append(names, name)
		}
		if err := s.supplierRepo.GetByNames(names, &sups); err == nil {
			for _, s := range sups {
				cache.suppliers[strings.ToLower(s.Name)] = s.ID
			}
		}
	}

	// Fetch only referenced Locations
	if len(uniqueLocations) > 0 {
		var locs []domain.Location
		names := make([]string, 0, len(uniqueLocations))
		for name := range uniqueLocations {
			names = append(names, name)
		}
		if err := s.locationRepo.GetByNames(names, &locs); err == nil {
			for _, l := range locs {
				cache.locations[strings.ToLower(l.Name)] = l.ID
			}
		}
	}

	return cache, nil
}
