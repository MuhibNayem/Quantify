package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"inventory/backend/internal/config"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"

	"github.com/sirupsen/logrus"
)

type ForecastingService interface {
	GenerateDemandForecast(productID *uint, periodInDays int) (*domain.DemandForecast, error)
	GetForecastDashboard() (map[string]interface{}, error)
	GetDemandForecastByID(id string) (*domain.DemandForecast, error)
}

type forecastingService struct {
	repo repository.ForecastingRepository
	cfg  *config.Config
}

func NewForecastingService(repo repository.ForecastingRepository, cfg *config.Config) ForecastingService {
	return &forecastingService{repo: repo, cfg: cfg}
}

func (s *forecastingService) GenerateDemandForecast(productID *uint, periodInDays int) (*domain.DemandForecast, error) {
	var products []domain.Product

	if productID != nil {
		// Use AI Service for single product forecast
		if s.cfg.AIServiceURL != "" {
			return s.generateAIForecast(*productID, periodInDays)
		}

		// Fallback to local logic if AI service not configured
		product, err := s.repo.GetProduct(*productID)
		if err != nil {
			return nil, fmt.Errorf("failed to get product: %w", err)
		}
		products = append(products, *product)
		// For local logic, we just return the last one generated or nil if batch
		// This is a bit hacky for batch, but fine for single product
		return s.processForecastForProducts(products, periodInDays)
	}

	// Batch processing for all products
	offset := 0
	limit := 100
	for {
		batchProducts, err := s.repo.GetProductsBatch(offset, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get products batch: %w", err)
		}
		if len(batchProducts) == 0 {
			break
		}

		if _, err := s.processForecastForProducts(batchProducts, periodInDays); err != nil {
			logrus.Errorf("Error processing batch offset %d: %v", offset, err)
			// Continue to next batch? Or stop? Let's log and continue to try to process as much as possible.
		}

		offset += limit
	}

	return nil, nil
}

func (s *forecastingService) generateAIForecast(productID uint, periodInDays int) (*domain.DemandForecast, error) {
	url := fmt.Sprintf("%s/forecast", s.cfg.AIServiceURL)
	payload := map[string]interface{}{
		"product_id": productID,
		"days":       periodInDays,
	}

	jsonPayload, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to call AI service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI service returned status: %d", resp.StatusCode)
	}

	var result struct {
		PredictedDemand int     `json:"predicted_demand"`
		ConfidenceScore float64 `json:"confidence_score"`
		Reasoning       string  `json:"reasoning"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode AI response: %w", err)
	}

	forecast := domain.DemandForecast{
		ProductID:       productID,
		ForecastPeriod:  fmt.Sprintf("%d_DAYS", periodInDays),
		PredictedDemand: result.PredictedDemand,
		ConfidenceScore: result.ConfidenceScore,
		Reasoning:       result.Reasoning,
		GeneratedAt:     time.Now(),
	}

	if err := s.repo.CreateForecast(&forecast); err != nil {
		return nil, fmt.Errorf("failed to save AI forecast: %w", err)
	}

	logrus.Infof("Generated AI forecast for product %d: PredictedDemand=%d", productID, result.PredictedDemand)
	return &forecast, nil
}

func (s *forecastingService) processForecastForProducts(products []domain.Product, periodInDays int) (*domain.DemandForecast, error) {
	var lastForecast *domain.DemandForecast
	for _, product := range products {
		salesData, err := s.repo.GetSalesDataForForecast(product.ID, periodInDays)
		if err != nil {
			logrus.Errorf("Failed to get sales data for product %d: %v", product.ID, err)
			continue
		}

		// Map sales to dates to handle missing days (0 sales)
		salesMap := make(map[string]int)
		for _, sale := range salesData {
			dateStr := sale.AdjustedAt.Format("2006-01-02")
			salesMap[dateStr] += sale.Quantity
		}

		var weightedSum float64
		var weightSum float64

		// Iterate through each day of the period
		now := time.Now()
		for i := 0; i < periodInDays; i++ {
			// Calculate date: start from (now - period) + i
			// actually, usually weighted average gives more weight to recent.
			// So let's iterate i from 1 to periodInDays.
			// Day 1 = (now - period) + 1 ... Day N = now.
			// Let's align:
			// targetDate = now.AddDate(0, 0, -periodInDays + i + 1)

			// Let's say period is 30 days.
			// i=0 -> weight=1. Date = now - 29 days.
			// ...
			// i=29 -> weight=30. Date = now.

			targetDate := now.AddDate(0, 0, -periodInDays+i+1)
			dateStr := targetDate.Format("2006-01-02")

			quantity := salesMap[dateStr] // 0 if missing

			weight := float64(i + 1)
			weightedSum += float64(quantity) * weight
			weightSum += weight
		}

		var predictedDemand int
		if weightSum > 0 {
			dailyDemand := weightedSum / weightSum
			predictedDemand = int(dailyDemand * float64(periodInDays))
		} else {
			predictedDemand = 0
		}

		forecast := domain.DemandForecast{
			ProductID:       product.ID,
			ForecastPeriod:  fmt.Sprintf("%d_DAYS", periodInDays),
			PredictedDemand: predictedDemand,
			GeneratedAt:     time.Now(),
		}
		if err := s.repo.CreateForecast(&forecast); err != nil {
			logrus.Errorf("Failed to save demand forecast for product %d: %v", product.ID, err)
			continue
		}
		logrus.Infof("Generated forecast for product %d: PredictedDemand=%d", product.ID, predictedDemand)
		lastForecast = &forecast
	}

	return lastForecast, nil
}

func (s *forecastingService) GetForecastDashboard() (map[string]interface{}, error) {
	topForecasts, err := s.repo.GetTopForecasts(10)
	if err != nil {
		return nil, fmt.Errorf("failed to get top forecasts: %w", err)
	}

	lowStock, err := s.repo.GetLowStockPredictions(10)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock predictions: %w", err)
	}

	return map[string]interface{}{
		"topForecasts": topForecasts,
		"lowStock":     lowStock,
	}, nil
}

func (s *forecastingService) GetDemandForecastByID(id string) (*domain.DemandForecast, error) {
	return s.repo.GetForecast(id)
}
