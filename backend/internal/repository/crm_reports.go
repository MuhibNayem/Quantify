package repository

import (
	"database/sql"
	"inventory/backend/internal/domain"
	"log"
	"time"
)

type CustomerInsight struct {
	UserID             uint
	Username           string
	FullName           string
	TotalSpent         float64
	GrossRevenue       float64
	TotalRefunded      float64
	NetRevenue         float64
	OrderCount         int
	LastOrderDate      *time.Time
	DaysSinceLastOrder int
}

func (r *reportsRepository) GetCustomerInsightsReport(startDate, endDate time.Time) ([]CustomerInsight, error) {
	log.Printf("DEBUG: Executing GetCustomerInsightsReport (New Logic) for range %v to %v", startDate, endDate)
	var insights []CustomerInsight

	query := `
		SELECT
			u.id,
			u.username,
			CASE
				WHEN u.first_name IS NOT NULL AND u.first_name != '' THEN u.first_name || ' ' || COALESCE(u.last_name, '')
				ELSE u.username
			END as full_name,
			COALESCE(SUM(o.total_amount), 0) as gross_revenue,
			COALESCE((
				SELECT SUM(ri.quantity * oi.unit_price)
				FROM returns ret
				JOIN return_items ri ON ri.return_id = ret.id
				JOIN order_items oi ON oi.id = ri.order_item_id
				JOIN orders ro ON ro.id = oi.order_id
				WHERE ro.customer_id = u.id
				AND ret.status = 'APPROVED'
				AND ret.created_at BETWEEN ? AND ?
			), 0) as total_refunded,
			COALESCE(SUM(o.total_amount), 0) - COALESCE((
				SELECT SUM(ri.quantity * oi.unit_price)
				FROM returns ret
				JOIN return_items ri ON ri.return_id = ret.id
				JOIN order_items oi ON oi.id = ri.order_item_id
				JOIN orders ro ON ro.id = oi.order_id
				WHERE ro.customer_id = u.id
				AND ret.status = 'APPROVED'
				AND ret.created_at BETWEEN ? AND ?
			), 0) as net_revenue,
			COUNT(o.id) as order_count,
			MAX(o.order_date) as last_order_date
		FROM users u
		JOIN orders o ON o.customer_id = u.id
		WHERE o.status = 'COMPLETED'
		AND o.order_date BETWEEN ? AND ?
		GROUP BY u.id, u.username, u.first_name, u.last_name
		ORDER BY net_revenue DESC
		LIMIT 50
	`

	rows, err := r.DB.Raw(query, startDate, endDate, startDate, endDate, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	for rows.Next() {
		var ci CustomerInsight
		var lastOrder sql.NullString

		if err := rows.Scan(&ci.UserID, &ci.Username, &ci.FullName, &ci.GrossRevenue, &ci.TotalRefunded, &ci.NetRevenue, &ci.OrderCount, &lastOrder); err != nil {
			return nil, err
		}

		ci.TotalSpent = ci.NetRevenue

		if lastOrder.Valid {
			if t, err := parseFlexibleTime(lastOrder.String); err == nil {
				ci.LastOrderDate = &t
				ci.DaysSinceLastOrder = int(now.Sub(t).Hours() / 24)
			} else {
				ci.DaysSinceLastOrder = -1
			}
		} else {
			ci.DaysSinceLastOrder = -1
		}

		insights = append(insights, ci)
	}

	return insights, nil
}

type ProductPerformanceAnalytics struct {
	ProductID    uint
	ProductName  string
	SupplierID   uint
	SupplierName string
	TotalRevenue float64
	TotalCost    float64
	TotalProfit  float64
	CurrentStock float64
	DaysOfStock  float64
}

func (r *reportsRepository) GetProductPerformanceAnalytics(startDate, endDate time.Time, supplierName string, minStock int) ([]ProductPerformanceAnalytics, error) {
	var results []ProductPerformanceAnalytics

	query := r.DB.Table("products").
		Select(`
			products.id as product_id,
			products.name as product_name,
			suppliers.id as supplier_id,
			suppliers.name as supplier_name,
			COALESCE(SUM(sa.quantity), 0) as total_sold_qty,
			COALESCE(SUM(sa.quantity * products.selling_price), 0) as total_revenue,
			COALESCE(SUM(sa.quantity * products.purchase_price), 0) as total_cost
		`).
		Joins("JOIN suppliers ON suppliers.id = products.supplier_id").
		Joins("LEFT JOIN stock_adjustments sa ON sa.product_id = products.id AND sa.type = 'STOCK_OUT' AND sa.reason_code = 'SALE' AND sa.adjusted_at BETWEEN ? AND ?", startDate, endDate).
		Group("products.id, products.name, suppliers.id, suppliers.name")

	if supplierName != "" {
		query = query.Where("suppliers.name ILIKE ?", "%"+supplierName+"%")
	}

	rows, err := query.Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results = []ProductPerformanceAnalytics{}

	for rows.Next() {
		var p ProductPerformanceAnalytics
		var totalSoldQty float64

		if err := rows.Scan(&p.ProductID, &p.ProductName, &p.SupplierID, &p.SupplierName, &totalSoldQty, &p.TotalRevenue, &p.TotalCost); err != nil {
			return nil, err
		}

		p.TotalProfit = p.TotalRevenue - p.TotalCost

		var currentStock float64
		r.DB.Model(&domain.Batch{}).Where("product_id = ?", p.ProductID).Select("COALESCE(SUM(quantity), 0)").Scan(&currentStock)
		p.CurrentStock = currentStock

		daysInPeriod := endDate.Sub(startDate).Hours() / 24
		if daysInPeriod < 1 {
			daysInPeriod = 1
		}

		dailySalesVelocity := totalSoldQty / daysInPeriod

		if dailySalesVelocity > 0 {
			p.DaysOfStock = p.CurrentStock / dailySalesVelocity
		} else {
			p.DaysOfStock = 999
		}

		results = append(results, p)
	}

	return results, nil
}

type ShrinkageItem struct {
	ProductID   uint
	ProductName string
	Reason      string
	Quantity    float64
	LostValue   float64
}

func (r *reportsRepository) GetShrinkageReport(startDate, endDate time.Time) ([]ShrinkageItem, error) {
	var items []ShrinkageItem

	query := `
		SELECT 
			p.id, 
			p.name, 
			sa.reason_code, 
			SUM(sa.quantity) as quantity,
			SUM(sa.quantity * p.purchase_price) as lost_value
		FROM stock_adjustments sa
		JOIN products p ON p.id = sa.product_id
		WHERE sa.type = 'STOCK_OUT' 
		AND sa.reason_code != 'SALE'
		AND sa.adjusted_at BETWEEN ? AND ?
		GROUP BY p.id, p.name, sa.reason_code
		ORDER BY lost_value DESC
	`

	rows, err := r.DB.Raw(query, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item ShrinkageItem
		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.Reason, &item.Quantity, &item.LostValue); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
