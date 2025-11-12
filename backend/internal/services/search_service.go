package services

import (
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"

	"gorm.io/gorm"
)

// SearchService orchestrates the global search functionality.
type SearchService struct {
	db             *gorm.DB
	searchRepo     *repository.SearchRepository
	productRepo    *repository.ProductRepository
	userRepo       *repository.UserRepository
	supplierRepo   *repository.SupplierRepository
	categoryRepo   *repository.CategoryRepository
}

// NewSearchService creates a new SearchService.
func NewSearchService(db *gorm.DB, searchRepo *repository.SearchRepository, productRepo *repository.ProductRepository, userRepo *repository.UserRepository, supplierRepo *repository.SupplierRepository, categoryRepo *repository.CategoryRepository) *SearchService {
	return &SearchService{
		db:             db,
		searchRepo:     searchRepo,
		productRepo:    productRepo,
		userRepo:       userRepo,
		supplierRepo:   supplierRepo,
		categoryRepo:   categoryRepo,
	}
}

// Search performs a global search and rehydrates the results with the full entity data.
func (s *SearchService) Search(query string) ([]repository.SearchResult, error) {
	// Get the raw search results (entity type and ID) from the search repository
	rawResults, err := s.searchRepo.Search(query)
	if err != nil {
		return nil, err
	}

	if len(rawResults) == 0 {
		return []repository.SearchResult{}, nil
	}

	// Create maps to hold the IDs for each entity type
	productIDs := []uint{}
	userIDs := []uint{}
	supplierIDs := []uint{}
	categoryIDs := []uint{}

	for _, res := range rawResults {
		switch res.EntityType {
		case "product":
			productIDs = append(productIDs, res.EntityID)
		case "user":
			userIDs = append(userIDs, res.EntityID)
		case "supplier":
			supplierIDs = append(supplierIDs, res.EntityID)
		case "category":
			categoryIDs = append(categoryIDs, res.EntityID)
		}
	}

	// Fetch the full entities in bulk to avoid N+1 queries
	products := make(map[uint]domain.Product)
	if len(productIDs) > 0 {
		var productList []domain.Product
		if err := s.db.Where(productIDs).Find(&productList).Error; err == nil {
			for _, p := range productList {
				products[p.ID] = p
			}
		}
	}

	users := make(map[uint]domain.User)
	if len(userIDs) > 0 {
		var userList []domain.User
		if err := s.db.Where(userIDs).Find(&userList).Error; err == nil {
			for _, u := range userList {
				u.Password = "" // Never return password
				users[u.ID] = u
			}
		}
	}

	suppliers := make(map[uint]domain.Supplier)
	if len(supplierIDs) > 0 {
		var supplierList []domain.Supplier
		if err := s.db.Where(supplierIDs).Find(&supplierList).Error; err == nil {
			for _, sup := range supplierList {
				suppliers[sup.ID] = sup
			}
		}
	}

	categories := make(map[uint]domain.Category)
	if len(categoryIDs) > 0 {
		var categoryList []domain.Category
		if err := s.db.Where(categoryIDs).Find(&categoryList).Error; err == nil {
			for _, cat := range categoryList {
				categories[cat.ID] = cat
			}
		}
	}

	// Rehydrate the final results with the fetched entities
	hydratedResults := make([]repository.SearchResult, 0, len(rawResults))
	for _, res := range rawResults {
		var entity interface{}
		found := false

		switch res.EntityType {
		case "product":
			entity, found = products[res.EntityID]
		case "user":
			entity, found = users[res.EntityID]
		case "supplier":
			entity, found = suppliers[res.EntityID]
		case "category":
			entity, found = categories[res.EntityID]
		}

		if found {
			res.Entity = entity
			hydratedResults = append(hydratedResults, res)
		}
	}

	return hydratedResults, nil
}
