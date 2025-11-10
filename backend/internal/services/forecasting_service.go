package services

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
)

type ForecastingService interface {
	GenerateDemandForecast(productID *uint, periodInDays int) error
}

type forecastingService struct {
	repo repository.ForecastingRepository
}

func NewForecastingService(repo repository.ForecastingRepository) ForecastingService {
	return &forecastingService{repo: repo}
}

func (s *forecastingService) GenerateDemandForecast(productID *uint, periodInDays int) error {
	var products []domain.Product
	var err error

	if productID != nil {
		product, err := s.repo.GetProduct(*productID)
		if err != nil {
			return fmt.Errorf("failed to get product: %w", err)
		}
		products = append(products, *product)
	} else {
		products, err = s.repo.GetAllProducts()
		if err != nil {
			return fmt.Errorf("failed to get all products: %w", err)
		}
	}

	for _, product := range products {
		salesData, err := s.repo.GetSalesDataForForecast(product.ID, periodInDays)
		if err != nil {
			logrus.Errorf("Failed to get sales data for product %d: %v", product.ID, err)
			continue
		}

		var weightedSum float64
		var weightSum float64
		for i, sale := range salesData {
			weight := float64(i + 1)
			weightedSum += float64(sale.Quantity) * weight
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
	}

	return nil
}
