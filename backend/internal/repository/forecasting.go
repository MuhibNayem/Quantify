package repository

import (
	"inventory/backend/internal/domain"
	"time"

	"gorm.io/gorm"
)

type ForecastingRepository interface {
	GetProduct(id uint) (*domain.Product, error)
	GetAllProducts() ([]domain.Product, error)
	GetSalesDataForForecast(productID uint, days int) ([]domain.StockAdjustment, error)
	CreateForecast(forecast *domain.DemandForecast) error
	GetProductsBatch(offset, limit int) ([]domain.Product, error)
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
	var sales []domain.StockAdjustment
	startDate := time.Now().AddDate(0, 0, -days)

	err := r.db.Where("product_id = ?", productID).
		Where("type = ?", "STOCK_OUT").
		Where("reason_code = ?", "SALE").
		Where("adjusted_at >= ?", startDate).
		Order("adjusted_at ASC").
		Find(&sales).Error

	if err != nil {
		return nil, err
	}
	return sales, nil
}

func (r *forecastingRepository) CreateForecast(forecast *domain.DemandForecast) error {
	return r.db.Create(forecast).Error
}

func (r *forecastingRepository) GetProductsBatch(offset, limit int) ([]domain.Product, error) {
	var products []domain.Product
	if err := r.db.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}
