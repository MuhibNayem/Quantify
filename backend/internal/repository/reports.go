package repository

import (
	"database/sql"
	"inventory/backend/internal/domain"
	"time"
)

type SalesTrend struct {
	Date       time.Time
	TotalSales float64
}

type TopSellingProduct struct {
	ProductID uint
	Name      string
	TotalSold float64
}

func GetSalesTrends(startDate, endDate time.Time, categoryID, locationID *uint) ([]SalesTrend, []TopSellingProduct, error) {
	var salesTrends []SalesTrend
	var topSellingProducts []TopSellingProduct

	// Sales Trends
	query := DB.Model(&domain.StockAdjustment{}).
		Select("DATE(adjusted_at) as date, SUM(quantity) as total_sales").
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

	err := query.Group("DATE(adjusted_at)").Order("DATE(adjusted_at)").Scan(&salesTrends).Error
	if err != nil {
		return nil, nil, err
	}

	// Top Selling Products
	query = DB.Model(&domain.StockAdjustment{}).
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

func GetInventoryTurnover(startDate, endDate time.Time, categoryID, locationID *uint) (float64, float64, error) {
	var costOfGoodsSold sql.NullFloat64

	// Calculate Cost of Goods Sold (COGS)
	query := DB.Model(&domain.StockAdjustment{}).
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
	var endInventoryValue sql.NullFloat64

	// This is a simplified calculation. A real implementation would need to reconstruct the inventory at a specific point in time.
	// For simplicity, we'll use the current inventory as a proxy for the end inventory and calculate the start inventory based on adjustments.

	// End Inventory Value
	query = DB.Model(&domain.Product{}).
		Select("SUM(products.selling_price * batches.quantity)").
		Joins("JOIN batches ON batches.product_id = products.id")
	if categoryID != nil {
		query = query.Where("products.category_id = ?", *categoryID)
	}
	if locationID != nil {
		query = query.Where("batches.location_id = ?", *locationID)
	}
	err = query.Scan(&endInventoryValue).Error
	if err != nil {
		return 0, 0, err
	}

	endInvValue := 0.0
	if endInventoryValue.Valid {
		endInvValue = endInventoryValue.Float64
	}

	// Start Inventory Value (End Inventory Value - changes during the period)
	var netChangeInValue sql.NullFloat64
	query = DB.Model(&domain.StockAdjustment{}).
		Select("SUM(CASE WHEN stock_adjustments.type = 'STOCK_IN' THEN stock_adjustments.quantity * products.purchase_price ELSE -stock_adjustments.quantity * products.purchase_price END)").
		Joins("JOIN products ON products.id = stock_adjustments.product_id").
		Where("stock_adjustments.adjusted_at BETWEEN ? AND ?", startDate, endDate)
	if categoryID != nil {
		query = query.Where("products.category_id = ?", *categoryID)
	}
	if locationID != nil {
		query = query.Where("stock_adjustments.location_id = ?", *locationID)
	}
	err = query.Scan(&netChangeInValue).Error
	if err != nil {
		return 0, 0, err
	}

	netChangeValue := 0.0
	if netChangeInValue.Valid {
		netChangeValue = netChangeInValue.Float64
	}

	startInvValue := endInvValue - netChangeValue

	averageInventoryValue := (startInvValue + endInvValue) / 2

	return cogsValue, averageInventoryValue, nil
}

func GetProfitMargin(startDate, endDate time.Time, categoryID, locationID *uint) (float64, float64, error) {
	var totalRevenue sql.NullFloat64
	var totalCost sql.NullFloat64

	// Calculate Total Revenue
	query := DB.Model(&domain.StockAdjustment{}).
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
	query = DB.Model(&domain.StockAdjustment{}).
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

func GetDailySalesSummary(date time.Time) (float64, error) {
	var totalSales sql.NullFloat64

	query := DB.Model(&domain.StockAdjustment{}).
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
