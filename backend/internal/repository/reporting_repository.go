package repository

import (
	"database/sql"
	"fmt"
	"inventory/backend/internal/domain"
	"time"

	"gorm.io/gorm"
)

type ReportsRepository struct {
	DB *gorm.DB
}

func NewReportsRepository(db *gorm.DB) *ReportsRepository {
	return &ReportsRepository{DB: db}
}

type SalesTrend struct {
	Date       time.Time
	TotalSales float64
}

type TopSellingProduct struct {
	ProductID uint
	Name      string
	TotalSold float64
}

func (r *ReportsRepository) GetSalesTrends(startDate, endDate time.Time, categoryID, locationID *uint, groupBy string) ([]SalesTrend, []TopSellingProduct, error) {
	var salesTrends []SalesTrend
	var topSellingProducts []TopSellingProduct

	// Sales Trends
	dateTrunc := "DATE"
	switch groupBy {
	case "week":
		dateTrunc = "WEEK"
	case "month":
		dateTrunc = "MONTH"
	}

	query := r.DB.Model(&domain.StockAdjustment{}).
		Select(fmt.Sprintf("%s(adjusted_at) as date, SUM(quantity) as total_sales", dateTrunc)).
		Where("type = ?", "STOCK_OUT").
		Where("reason_code = ?", "SALE").
		Where("adjusted_at BETWEEN ? AND ?", startDate, endDate)

	if categoryID != nil {
		query = query.Joins("JOIN products ON products.id = stock_adjustments.product_id").
			Where("products.category_id = ?", *categoryID)
	}
	if locationID != nil {
		query = query.Where("location_id = ?", *locationID)
	}

	err := query.Group(fmt.Sprintf("%s(adjusted_at)", dateTrunc)).Order(fmt.Sprintf("%s(adjusted_at)", dateTrunc)).Scan(&salesTrends).Error
	if err != nil {
		return nil, nil, err
	}

	// Top Selling Products
	query = r.DB.Model(&domain.StockAdjustment{}).
		Select("products.id as product_id, products.name, SUM(stock_adjustments.quantity) as total_sold").
		Joins("JOIN products ON products.id = stock_adjustments.product_id").
		Where("stock_adjustments.type = ?", "STOCK_OUT").
		Where("stock_adjustments.reason_code = ?", "SALE").
		Where("stock_adjustments.adjusted_at BETWEEN ? AND ?", startDate, endDate)

	if categoryID != nil {
		query = query.Where("products.category_id = ?", *categoryID)
	}
	if locationID != nil {
		query = query.Where("stock_adjustments.location_id = ?", *locationID)
	}

	err = query.Group("products.id, products.name").Order("total_sold DESC").Limit(10).Scan(&topSellingProducts).Error
	if err != nil {
		return nil, nil, err
	}

	return salesTrends, topSellingProducts, nil
}

func (r *ReportsRepository) GetInventoryTurnover(startDate, endDate time.Time, categoryID, locationID *uint) (float64, float64, error) {
	var costOfGoodsSold sql.NullFloat64

	// Calculate Cost of Goods Sold (COGS)
	query := r.DB.Model(&domain.StockAdjustment{}).
		Select("SUM(stock_adjustments.quantity * products.purchase_price)").
		Joins("JOIN products ON products.id = stock_adjustments.product_id").
		Where("stock_adjustments.type = ?", "STOCK_OUT").
		Where("stock_adjustments.reason_code = ?", "SALE").
		Where("stock_adjustments.adjusted_at BETWEEN ? AND ?", startDate, endDate)

	if categoryID != nil {
		query = query.Where("products.category_id = ?", *categoryID)
	}
	if locationID != nil {
		query = query.Where("stock_adjustments.location_id = ?", *locationID)
	}

	err := query.Scan(&costOfGoodsSold).Error
	if err != nil {
		return 0, 0, err
	}

	cogsValue := 0.0
	if costOfGoodsSold.Valid {
		cogsValue = costOfGoodsSold.Float64
	}

	// Calculate Average Inventory Value
	startInvValue, err := r.getInventoryValueAt(startDate, categoryID, locationID)
	if err != nil {
		return 0, 0, err
	}
	endInvValue, err := r.getInventoryValueAt(endDate, categoryID, locationID)
	if err != nil {
		return 0, 0, err
	}

	averageInventoryValue := (startInvValue + endInvValue) / 2

	return cogsValue, averageInventoryValue, nil
}

func (r *ReportsRepository) getInventoryValueAt(date time.Time, categoryID, locationID *uint) (float64, error) {
	var totalValue sql.NullFloat64

	// Get total stock value up to the given date
	query := r.DB.Model(&domain.StockAdjustment{}).
		Select("SUM(CASE WHEN type = 'STOCK_IN' THEN quantity * products.purchase_price ELSE -quantity * products.purchase_price END)").
		Joins("JOIN products ON products.id = stock_adjustments.product_id").
		Where("adjusted_at <= ?", date)

	if categoryID != nil {
		query = query.Where("products.category_id = ?", *categoryID)
	}
	if locationID != nil {
		query = query.Where("location_id = ?", *locationID)
	}

	err := query.Scan(&totalValue).Error
	if err != nil {
		return 0, err
	}

	value := 0.0
	if totalValue.Valid {
		value = totalValue.Float64
	}

	return value, nil
}

func (r *ReportsRepository) GetProfitMargin(startDate, endDate time.Time, categoryID, locationID *uint) (float64, float64, error) {
	var totalRevenue sql.NullFloat64
	var totalCost sql.NullFloat64

	// Calculate Total Revenue
	query := r.DB.Model(&domain.StockAdjustment{}).
		Select("SUM(stock_adjustments.quantity * products.selling_price)").
		Joins("JOIN products ON products.id = stock_adjustments.product_id").
		Where("stock_adjustments.type = ?", "STOCK_OUT").
		Where("stock_adjustments.reason_code = ?", "SALE").
		Where("stock_adjustments.adjusted_at BETWEEN ? AND ?", startDate, endDate)

	if categoryID != nil {
		query = query.Where("products.category_id = ?", *categoryID)
	}
	if locationID != nil {
		query = query.Where("stock_adjustments.location_id = ?", *locationID)
	}

	err := query.Scan(&totalRevenue).Error
	if err != nil {
		return 0, 0, err
	}

	revenueValue := 0.0
	if totalRevenue.Valid {
		revenueValue = totalRevenue.Float64
	}

	// Calculate Total Cost (COGS)
	query = r.DB.Model(&domain.StockAdjustment{}).
		Select("SUM(stock_adjustments.quantity * products.purchase_price)").
		Joins("JOIN products ON products.id = stock_adjustments.product_id").
		Where("stock_adjustments.type = ?", "STOCK_OUT").
		Where("stock_adjustments.reason_code = ?", "SALE").
		Where("stock_adjustments.adjusted_at BETWEEN ? AND ?", startDate, endDate)

	if categoryID != nil {
		query = query.Where("products.category_id = ?", *categoryID)
	}
	if locationID != nil {
		query = query.Where("stock_adjustments.location_id = ?", *locationID)
	}

	err = query.Scan(&totalCost).Error
	if err != nil {
		return 0, 0, err
	}

	costValue := 0.0
	if totalCost.Valid {
		costValue = totalCost.Float64
	}

	return revenueValue, costValue, nil
}

func (r *ReportsRepository) GetDailySalesSummary(date time.Time) (float64, error) {
	var totalSales sql.NullFloat64

	query := r.DB.Model(&domain.StockAdjustment{}).
		Select("SUM(stock_adjustments.quantity * products.selling_price)").
		Joins("JOIN products ON products.id = stock_adjustments.product_id").
		Where("stock_adjustments.type = ?", "STOCK_OUT").
		Where("stock_adjustments.reason_code = ?", "SALE").
		Where("DATE(stock_adjustments.adjusted_at) = ?", date.Format("2006-01-02"))

	err := query.Scan(&totalSales).Error
	if err != nil {
		return 0, err
	}

	salesValue := 0.0
	if totalSales.Valid {
		salesValue = totalSales.Float64
	}

	return salesValue, nil
}
