package services

import (
	"fmt"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
	"time"

	"github.com/sirupsen/logrus"
)

type ReplenishmentService interface {
	GenerateReorderSuggestions() error
}

type replenishmentService struct {
	repo repository.ReplenishmentRepository
}

func NewReplenishmentService(repo repository.ReplenishmentRepository) ReplenishmentService {
	return &replenishmentService{repo: repo}
}

func (s *replenishmentService) GenerateReorderSuggestions() error {
	settings, err := s.repo.GetAllProductAlertSettings()
	if err != nil {
		return fmt.Errorf("failed to fetch alert settings: %w", err)
	}

	if len(settings) == 0 {
		return nil
	}

	// Collect all product IDs
	productIDs := make([]uint, len(settings))
	for i, setting := range settings {
		productIDs[i] = setting.ProductID
	}

	// Bulk fetch data
	stockMap, err := s.repo.GetStockLevels(productIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch stock levels: %w", err)
	}

	pendingSuggestions, err := s.repo.GetPendingSuggestionsMap(productIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch pending suggestions: %w", err)
	}

	pendingPOs, err := s.repo.GetPendingPOsMap(productIDs)
	if err != nil {
		return fmt.Errorf("failed to fetch pending POs: %w", err)
	}

	for _, setting := range settings {
		currentStock := stockMap[setting.ProductID] // Default 0 if not found

		if currentStock <= setting.LowStockLevel {
			// Check for pending suggestions
			if pendingSuggestions[setting.ProductID] {
				continue // Already suggested
			}

			// Check for pending POs
			if pendingPOs[setting.ProductID] {
				continue // Already ordered
			}

			// Create Suggestion
			// Calculate quantity: Target (OverStockLevel) - Current
			// If OverStockLevel is 0 or less than LowStock, use a default or just LowStock * 2
			targetStock := setting.OverStockLevel
			if targetStock <= setting.LowStockLevel {
				targetStock = setting.LowStockLevel * 3 // Default heuristic
			}
			suggestedQty := targetStock - currentStock
			if suggestedQty <= 0 {
				suggestedQty = 10 // Fallback
			}

			// We need the supplier ID. Since we preloaded Product in GetAllProductAlertSettings,
			// we might have it if Product has SupplierID.
			// Let's assume Product is preloaded in settings.
			// If not, we might need to fetch it. But wait, GetAllProductAlertSettings preloads Product.
			// Does Product have SupplierID? Yes.
			// Does it have the full Supplier object? Maybe not fully preloaded if nested.
			// But we only need SupplierID for the suggestion.

			supplierID := setting.Product.SupplierID
			if supplierID == 0 {
				logrus.Warnf("Product %d has no supplier configured, skipping suggestion", setting.ProductID)
				continue
			}

			suggestion := &domain.ReorderSuggestion{
				ProductID:              setting.ProductID,
				SupplierID:             supplierID,
				CurrentStock:           currentStock,
				PredictedDemand:        0, // TODO: Integrate with forecasting
				SuggestedOrderQuantity: suggestedQty,
				LeadTimeDays:           7, // Default
				Status:                 "PENDING",
				SuggestedAt:            time.Now(),
			}

			if err := s.repo.CreateReorderSuggestion(suggestion); err != nil {
				logrus.Errorf("Failed to create suggestion for product %d: %v", setting.ProductID, err)
			} else {
				logrus.Infof("Created reorder suggestion for product %d, qty %d", setting.ProductID, suggestedQty)
			}
		}
	}

	return nil
}
