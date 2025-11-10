package repository

import (
	"inventory/backend/internal/domain"
	"gorm.io/gorm"
)

type ForecastingRepository interface {
	GetProduct(id uint) (*domain.Product, error)
	GetAllProducts() ([]domain.Product, error)
	GetSalesDataForForecast(productID uint, days int) ([]domain.StockAdjustment, error)
	CreateForecast(forecast *domain.DemandForecast) error
}

type forecastingRepository struct {
	db *gorm.DB
}

func NewForecastingRepository(db *gorm.DB) ForecastingRepository {
	return &forecastingRepository{db: db}
}

func (r *forecastingRepository) GetProduct(id uint) (*domain.Product, error) {
	var product domain.Product
	if err := r.db.First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *forecastingRepository) GetAllProducts() ([]domain.Product, error) {
	var products []domain.Product
	if err := r.db.Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *forecastingRepository) GetSalesDataForForecast(productID uint, days int) ([]domain.StockAdjustment, error) {
	// This is a mock implementation. In a real application, you would query your sales data.
	// For now, we'll return an empty slice.
	return []domain.StockAdjustment{}, nil
}

func (r *forecastingRepository) CreateForecast(forecast *domain.DemandForecast) error {
	return r.db.Create(forecast).Error
}
