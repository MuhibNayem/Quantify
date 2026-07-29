package repository

import (
	"database/sql"
	"fmt"
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

func (r *reportsRepository) GetSalesTrends(startDate, endDate time.Time, categoryID, locationID, productID *uint, groupBy, timezone string) ([]SalesTrend, []TopSellingProduct, error) {
	var salesTrends []SalesTrend
	var topSellingProducts []TopSellingProduct

	if !isValidTimezone(timezone) {
		timezone = "UTC"
	}

	var dateExpr string
	switch groupBy {
	case "iso_week", "weekly", "week":
		dateExpr = fmt.Sprintf("DATE_TRUNC('week', adjusted_at AT TIME ZONE '%s')", timezone)
	case "calendar_week":
		dateExpr = fmt.Sprintf("(DATE_TRUNC('week', (adjusted_at AT TIME ZONE '%s') + INTERVAL '1 day') - INTERVAL '1 day')", timezone)
	case "daily", "day":
		dateExpr = fmt.Sprintf("DATE_TRUNC('day', adjusted_at AT TIME ZONE '%s')", timezone)
	case "monthly", "month":
		dateExpr = fmt.Sprintf("DATE_TRUNC('month', adjusted_at AT TIME ZONE '%s')", timezone)
	default:
		return nil, nil, fmt.Errorf("unsupported groupBy: %s", groupBy)
	}

	query := r.DB.Model(&domain.StockAdjustment{}).
		Select(fmt.Sprintf("%s as date, SUM(quantity) as total_sales", dateExpr)).
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
	if productID != nil {
		query = query.Where("stock_adjustments.product_id = ?", *productID)
	}

	err := query.Group(dateExpr).Order(dateExpr).Scan(&salesTrends).Error
	if err != nil {
		return nil, nil, err
	}

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
	if productID != nil {
		query = query.Where("stock_adjustments.product_id = ?", *productID)
	}

	err = query.Group("products.id, products.name").Order("total_sold DESC").Limit(10).Scan(&topSellingProducts).Error
	if err != nil {
		return nil, nil, err
	}

	return salesTrends, topSellingProducts, nil
}

type HeatmapPoint struct {
	DayOfWeek  int
	HourOfDay  int
	TotalSales float64
}

func (r *reportsRepository) GetHourlySalesHeatmap(startDate, endDate time.Time, timezone string) ([]HeatmapPoint, error) {
	var points []HeatmapPoint

	if !isValidTimezone(timezone) {
		timezone = "UTC"
	}

	query := `
		SELECT
			EXTRACT(DOW FROM adjusted_at AT TIME ZONE ?) as day_of_week,
			EXTRACT(HOUR FROM adjusted_at AT TIME ZONE ?) as hour_of_day,
			SUM(quantity * p.selling_price) as total_sales
		FROM stock_adjustments sa
		JOIN products p ON p.id = sa.product_id
		WHERE sa.type = 'STOCK_OUT'
		AND sa.reason_code = 'SALE'
		AND sa.adjusted_at BETWEEN ? AND ?
		GROUP BY day_of_week, hour_of_day
		ORDER BY day_of_week, hour_of_day
	`

	rows, err := r.DB.Raw(query, timezone, timezone, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p HeatmapPoint
		if err := rows.Scan(&p.DayOfWeek, &p.HourOfDay, &p.TotalSales); err != nil {
			return nil, err
		}
		points = append(points, p)
	}

	return points, nil
}

type BasketAnalysisItem struct {
	ProductA     uint
	ProductAName string
	ProductB     uint
	ProductBName string
	Frequency    int
}

func (r *reportsRepository) GetBasketAnalysisReport(startDate, endDate time.Time, page, limit int) ([]BasketAnalysisItem, int, error) {
	var items []BasketAnalysisItem

	offset := (page - 1) * limit

	countQuery := `
		SELECT COUNT(*) FROM (
			SELECT oi1.product_id, oi2.product_id
			FROM order_items oi1
			JOIN order_items oi2 ON oi1.order_id = oi2.order_id AND oi1.product_id < oi2.product_id
			JOIN products p1 ON p1.id = oi1.product_id
			JOIN products p2 ON p2.id = oi2.product_id
			JOIN orders o ON o.id = oi1.order_id
			WHERE o.order_date BETWEEN ? AND ?
			AND p1.deleted_at IS NULL
			AND p2.deleted_at IS NULL
			AND p1.name != ''
			AND p2.name != ''
			GROUP BY oi1.product_id, oi2.product_id
		) AS pair_counts
	`

	var total int
	if err := r.DB.Raw(countQuery, startDate, endDate).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			oi1.product_id as product_a,
			p1.name as product_a_name,
			oi2.product_id as product_b,
			p2.name as product_b_name,
			COUNT(*) as frequency
		FROM order_items oi1
		JOIN order_items oi2 ON oi1.order_id = oi2.order_id AND oi1.product_id < oi2.product_id
		JOIN products p1 ON p1.id = oi1.product_id
		JOIN products p2 ON p2.id = oi2.product_id
		JOIN orders o ON o.id = oi1.order_id
		WHERE o.order_date BETWEEN ? AND ?
		AND p1.deleted_at IS NULL
		AND p2.deleted_at IS NULL
		AND p1.name != ''
		AND p2.name != ''
		GROUP BY oi1.product_id, p1.name, oi2.product_id, p2.name
		ORDER BY frequency DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.DB.Raw(query, startDate, endDate, limit, offset).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var item BasketAnalysisItem
		if err := rows.Scan(&item.ProductA, &item.ProductAName, &item.ProductB, &item.ProductBName, &item.Frequency); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}

	return items, total, nil
}

type ReturnAnalysisItem struct {
	ProductID     uint
	ProductName   string
	Reason        string
	ReturnCount   int
	TotalRefunded float64
}

func (r *reportsRepository) GetCustomerReturnAnalysisReport(startDate, endDate time.Time) ([]ReturnAnalysisItem, error) {
	var items []ReturnAnalysisItem

	rows, err := r.DB.Raw(`
		SELECT 
			p.id, 
			p.name, 
			ri.reason, 
			COUNT(ri.id) as return_count,
			SUM(ri.quantity * p.selling_price) as total_refunded
		FROM return_items ri
		JOIN products p ON p.id = ri.product_id
		JOIN returns r ON r.id = ri.return_id
		WHERE r.status = 'COMPLETED'
		AND r.created_at BETWEEN ? AND ?
		GROUP BY p.id, p.name, ri.reason
		ORDER BY return_count DESC
	`, startDate, endDate).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item ReturnAnalysisItem
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.Reason, &item.ReturnCount, &item.TotalRefunded); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *reportsRepository) GetDailySalesSummary(date time.Time) (float64, error) {
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

type EmployeeSalesStats struct {
	UserID      uint
	Username    string
	TotalOrders int
	TotalSales  float64
}

func (r *reportsRepository) GetSalesByEmployeeReport(startDate, endDate time.Time) ([]EmployeeSalesStats, error) {
	var stats []EmployeeSalesStats

	query := `
		SELECT 
			u.id, 
			u.username, 
			COUNT(o.id) as total_orders, 
			SUM(o.total_amount) as total_sales
		FROM orders o
		JOIN users u ON u.id = o.user_id
		LEFT JOIN roles r ON r.id = u.role_id
		WHERE o.status = 'COMPLETED'
		AND (r.name != 'Customer' OR r.name IS NULL)
		AND o.order_date BETWEEN ? AND ?
		GROUP BY u.id, u.username
		ORDER BY total_sales DESC
	`

	rows, err := r.DB.Raw(query, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s EmployeeSalesStats
		if err := rows.Scan(&s.UserID, &s.Username, &s.TotalOrders, &s.TotalSales); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

type CategoryPerformance struct {
	CategoryID    uint
	CategoryName  string
	TotalSales    float64
	TotalCost     float64
	GrossMargin   float64
	MarginPercent float64
	ItemCount     float64
}

func (r *reportsRepository) GetCategoryDrillDownReport(startDate, endDate time.Time) ([]CategoryPerformance, error) {
	var results []CategoryPerformance

	query := `
		SELECT 
			c.id, 
			c.name, 
			SUM(sa.quantity * p.selling_price) as total_sales,
			SUM(sa.quantity * p.purchase_price) as total_cost,
			SUM(sa.quantity) as item_count
		FROM stock_adjustments sa
		JOIN products p ON p.id = sa.product_id
		JOIN categories c ON c.id = p.category_id
		WHERE sa.type = 'STOCK_OUT' 
		AND sa.reason_code = 'SALE'
		AND sa.adjusted_at BETWEEN ? AND ?
		GROUP BY c.id, c.name
		ORDER BY total_sales DESC
	`

	rows, err := r.DB.Raw(query, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cp CategoryPerformance
		if err := rows.Scan(&cp.CategoryID, &cp.CategoryName, &cp.TotalSales, &cp.TotalCost, &cp.ItemCount); err != nil {
			return nil, err
		}

		cp.GrossMargin = cp.TotalSales - cp.TotalCost
		if cp.TotalSales > 0 {
			cp.MarginPercent = (cp.GrossMargin / cp.TotalSales) * 100
		}

		results = append(results, cp)
	}

	return results, nil
}
