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
	GetTopForecasts(limit int) ([]ForecastDashboardItem, error)
	GetLowStockPredictions(limit int) ([]ForecastDashboardItem, error)
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

type ForecastDashboardItem struct {
	ProductID       uint
	ProductName     string
	PredictedDemand int
	CurrentStock    int
}

func (r *forecastingRepository) GetTopForecasts(limit int) ([]ForecastDashboardItem, error) {
	var items []ForecastDashboardItem
	// Get latest forecast for each product
	// Simplified: assuming one forecast per product per period, or we take the latest `generated_at`
	// We'll filter by forecasts generated in the last 24 hours to ensure freshness
	yesterday := time.Now().AddDate(0, 0, -1)

	err := r.db.Table("demand_forecasts").
		Select("products.id as product_id, products.name as product_name, demand_forecasts.predicted_demand, COALESCE(SUM(batches.quantity), 0) as current_stock").
		Joins("JOIN products ON products.id = demand_forecasts.product_id").
		Joins("LEFT JOIN batches ON batches.product_id = products.id").
		Where("demand_forecasts.generated_at >= ?", yesterday).
		Group("products.id, products.name, demand_forecasts.predicted_demand").
		Order("demand_forecasts.predicted_demand DESC").
		Limit(limit).
		Scan(&items).Error

	return items, err
}

func (r *forecastingRepository) GetLowStockPredictions(limit int) ([]ForecastDashboardItem, error) {
	var items []ForecastDashboardItem
	yesterday := time.Now().AddDate(0, 0, -1)

	err := r.db.Table("demand_forecasts").
		Select("products.id as product_id, products.name as product_name, demand_forecasts.predicted_demand, COALESCE(SUM(batches.quantity), 0) as current_stock").
		Joins("JOIN products ON products.id = demand_forecasts.product_id").
		Joins("LEFT JOIN batches ON batches.product_id = products.id").
		Where("demand_forecasts.generated_at >= ?", yesterday).
		Group("products.id, products.name, demand_forecasts.predicted_demand").
		Having("COALESCE(SUM(batches.quantity), 0) < demand_forecasts.predicted_demand").
		Order("demand_forecasts.predicted_demand DESC").
		Limit(limit).
		Scan(&items).Error

	return items, err
}
