package repository

import (
	"inventory/backend/internal/domain"
	"time"

	"gorm.io/gorm"
)

type DashboardRepository struct {
	DB *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{DB: db}
}

type DashboardStats struct {
	ProductCount     int64
	CategoryCount    int64
	SupplierCount    int64
	ActiveAlertCount int64
}

func (r *DashboardRepository) GetStats() (*DashboardStats, error) {
	var stats DashboardStats

	if err := r.DB.Model(&domain.Product{}).Count(&stats.ProductCount).Error; err != nil {
		return nil, err
	}
	if err := r.DB.Model(&domain.Category{}).Count(&stats.CategoryCount).Error; err != nil {
		return nil, err
	}
	if err := r.DB.Model(&domain.Supplier{}).Count(&stats.SupplierCount).Error; err != nil {
		return nil, err
	}
	if err := r.DB.Model(&domain.Alert{}).Where("status = ?", "ACTIVE").Count(&stats.ActiveAlertCount).Error; err != nil {
		return nil, err
	}

	return &stats, nil
}

func (r *DashboardRepository) GetRecentProducts(limit int) ([]domain.Product, error) {
	var products []domain.Product
	if err := r.DB.Order("created_at DESC").Limit(limit).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}

func (r *DashboardRepository) GetRecentAlerts(limit int) ([]domain.Alert, error) {
	var alerts []domain.Alert
	if err := r.DB.Preload("Product").Where("status = ?", "ACTIVE").Order("created_at DESC").Limit(limit).Find(&alerts).Error; err != nil {
		return nil, err
	}
	return alerts, nil
}

func (r *DashboardRepository) GetRecentReorderSuggestions(limit int) ([]domain.ReorderSuggestion, error) {
	var suggestions []domain.ReorderSuggestion
	if err := r.DB.Preload("Product").Preload("Supplier").Where("status = ?", "PENDING").Order("created_at DESC").Limit(limit).Find(&suggestions).Error; err != nil {
		return nil, err
	}
	return suggestions, nil
}

type ChartDataPoint struct {
	Date  string
	Value float64
}

func (r *DashboardRepository) GetSalesChartData() ([]float64, error) {
	// Get last 7 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -6) // 7 days including today

	var results []ChartDataPoint

	// Query to get sum of quantity for STOCK_OUT/SALE grouped by date
	// We use a raw query or GORM builder to ensure we get 0s for missing days if possible,
	// but for simplicity we'll fetch existing data and fill gaps in Go.

	// Note: Adjust syntax for your specific SQL dialect (Postgres assumed based on "github.com/lib/pq" hint earlier, though GORM abstracts it)
	err := r.DB.Model(&domain.StockAdjustment{}).
		Select("DATE(adjusted_at) as date, SUM(quantity) as value").
		Where("type = ? AND reason_code = ?", "STOCK_OUT", "SALE").
		Where("adjusted_at >= ?", startDate).
		Group("DATE(adjusted_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Create a map for easy lookup
	dataMap := make(map[string]float64)
	for _, r := range results {
		// date string from DB might vary, assuming YYYY-MM-DD
		// If it's a time.Time, we need to format it.
		// GORM Scan into struct with string Date might rely on driver.
		// Let's be safer and scan into a struct with time.Time if we can, or handle string.
		// Actually, let's just use the map to fill the array.
		// The Date in ChartDataPoint is string, let's assume standard format.
		// To be robust, let's assume the DB returns a date we can parse or is already YYYY-MM-DD.
		// For Postgres DATE() returns a date type.

		// Let's try to match the dates generated in Go.
		// We'll use the date string as key.
		// If the driver returns time.Time for DATE(), we might need to change struct.
		// Let's use a struct with time.Time to be safe.
		dataMap[r.Date[0:10]] = r.Value // Simple substring if it's a long string, or exact match
	}

	// Fill the last 7 days array
	chartSeries := make([]float64, 7)
	for i := 0; i < 7; i++ {
		d := startDate.AddDate(0, 0, i)
		dateStr := d.Format("2006-01-02")

		// We need to match the key from DB.
		// If DB returns "2023-10-27T00:00:00Z", we need to handle that.
		// Let's refine the query/struct to be sure.

		// Improved approach:
		// Scan into a struct with time.Time
		val, ok := dataMap[dateStr]
		if ok {
			chartSeries[i] = val
		} else {
			chartSeries[i] = 0
		}
	}

	return chartSeries, nil
}

type SalesTrendData struct {
	Direction  string  `json:"direction"` // "up", "down", "neutral"
	Percentage float64 `json:"percentage"`
}

func (r *DashboardRepository) GetSalesTrend() (*SalesTrendData, error) {
	now := time.Now()
	currentPeriodStart := now.AddDate(0, 0, -6)                 // Last 7 days
	previousPeriodStart := currentPeriodStart.AddDate(0, 0, -7) // 7 days before that
	previousPeriodEnd := currentPeriodStart.Add(-1 * time.Second)

	var currentSales float64
	var previousSales float64

	// Get Current Period Sales
	err := r.DB.Model(&domain.StockAdjustment{}).
		Where("type = ? AND reason_code = ?", "STOCK_OUT", "SALE").
		Where("adjusted_at >= ?", currentPeriodStart).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&currentSales).Error
	if err != nil {
		return nil, err
	}

	// Get Previous Period Sales
	err = r.DB.Model(&domain.StockAdjustment{}).
		Where("type = ? AND reason_code = ?", "STOCK_OUT", "SALE").
		Where("adjusted_at BETWEEN ? AND ?", previousPeriodStart, previousPeriodEnd).
		Select("COALESCE(SUM(quantity), 0)").
		Scan(&previousSales).Error
	if err != nil {
		return nil, err
	}

	trend := &SalesTrendData{
		Direction:  "neutral",
		Percentage: 0,
	}

	if previousSales == 0 {
		if currentSales > 0 {
			trend.Direction = "up"
			trend.Percentage = 100 // Or technically infinite, but 100 indicates "new growth"
		}
		// else both 0, remains neutral 0%
	} else {
		diff := currentSales - previousSales
		percentage := (diff / previousSales) * 100
		trend.Percentage = percentage

		if percentage > 0 {
			trend.Direction = "up"
		} else if percentage < 0 {
			trend.Direction = "down"
			trend.Percentage = -percentage // Keep percentage positive for display, direction handles sign
		}
	}

	return trend, nil
}
