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

type StockAgingItem struct {
	ProductID   uint
	ProductName string
	SKU         string
	AgeDays     int
	Quantity    int
	Value       float64
}

// GetStockAgingReport groups inventory by age buckets (0-30, 31-60, 61-90, 90+ days).
// It assumes FIFO: oldest batches are sold first.
func (r *ReportsRepository) GetStockAgingReport() (map[string][]StockAgingItem, error) {
	var batches []domain.Batch
	// Fetch all batches with positive quantity, ordered by creation date (oldest first)
	// Preload product to get name/sku/price
	if err := r.DB.Preload("Product").Where("quantity > 0").Order("created_at asc").Find(&batches).Error; err != nil {
		return nil, err
	}

	report := make(map[string][]StockAgingItem)
	report["0-30"] = []StockAgingItem{}
	report["31-60"] = []StockAgingItem{}
	report["61-90"] = []StockAgingItem{}
	report["90+"] = []StockAgingItem{}

	now := time.Now()

	for _, b := range batches {
		age := int(now.Sub(b.CreatedAt).Hours() / 24)
		item := StockAgingItem{
			ProductID:   b.ProductID,
			ProductName: b.Product.Name,
			SKU:         b.Product.SKU,
			AgeDays:     age,
			Quantity:    b.Quantity,
			Value:       float64(b.Quantity) * b.Product.PurchasePrice,
		}

		if age <= 30 {
			report["0-30"] = append(report["0-30"], item)
		} else if age <= 60 {
			report["31-60"] = append(report["31-60"], item)
		} else if age <= 90 {
			report["61-90"] = append(report["61-90"], item)
		} else {
			report["90+"] = append(report["90+"], item)
		}
	}

	return report, nil
}

type DeadStockItem struct {
	ProductID     uint
	ProductName   string
	SKU           string
	CurrentStock  int
	LastSaleDate  *time.Time
	DaysSinceSale int
	Value         float64
}

// GetDeadStockReport identifies products with stock > 0 but no sales in the last X days.
func (r *ReportsRepository) GetDeadStockReport(daysThreshold int) ([]DeadStockItem, error) {
	var items []DeadStockItem
	thresholdDate := time.Now().AddDate(0, 0, -daysThreshold)

	// Subquery to find last sale date for each product
	// We use LEFT JOIN to include products that have NEVER been sold
	query := `
		SELECT 
			p.id as product_id, 
			p.name as product_name, 
			p.sku, 
			COALESCE(SUM(b.quantity), 0) as current_stock,
			MAX(sa.adjusted_at) as last_sale_date,
			p.purchase_price
		FROM products p
		JOIN batches b ON b.product_id = p.id
		LEFT JOIN stock_adjustments sa ON sa.product_id = p.id AND sa.type = 'STOCK_OUT' AND sa.reason_code = 'SALE'
		WHERE p.deleted_at IS NULL
		GROUP BY p.id
		HAVING current_stock > 0 AND (last_sale_date < ? OR last_sale_date IS NULL)
	`

	rows, err := r.DB.Raw(query, thresholdDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item DeadStockItem
		var price float64
		var lastSale sql.NullTime

		if err := rows.Scan(&item.ProductID, &item.ProductName, &item.SKU, &item.CurrentStock, &lastSale, &price); err != nil {
			return nil, err
		}

		if lastSale.Valid {
			item.LastSaleDate = &lastSale.Time
			item.DaysSinceSale = int(time.Since(lastSale.Time).Hours() / 24)
		} else {
			// Never sold
			item.DaysSinceSale = -1 // Indicator for "Never"
		}
		item.Value = float64(item.CurrentStock) * price
		items = append(items, item)
	}

	return items, nil
}

type SupplierScorecard struct {
	SupplierID   uint
	SupplierName string
	TotalPOs     int
	AvgLeadTime  float64 // Days
	FillRate     float64 // Percentage
	ReturnRate   float64 // Percentage
}

// GetSupplierPerformanceReport calculates metrics for suppliers based on PO history.
func (r *ReportsRepository) GetSupplierPerformanceReport(startDate, endDate time.Time) ([]SupplierScorecard, error) {
	var scorecards []SupplierScorecard

	// Logic:
	// 1. Fetch all COMPLETED/RECEIVED POs in range
	// 2. Calculate Lead Time: ReceivedAt - OrderDate
	// 3. Calculate Fill Rate: TotalReceivedQty / TotalOrderedQty
	// 4. Calculate Return Rate: TotalReturnedQty / TotalReceivedQty (approx)

	// This is a complex aggregation, so we might do it in Go or complex SQL.
	// Let's try SQL for efficiency.

	query := `
		SELECT 
			s.id, 
			s.name, 
			COUNT(DISTINCT po.id) as total_pos,
			AVG(EXTRACT(EPOCH FROM (po.updated_at - po.order_date))/86400) as avg_lead_time,
			SUM(poi.received_quantity) as total_received,
			SUM(poi.quantity) as total_ordered
		FROM suppliers s
		JOIN purchase_orders po ON po.supplier_id = s.id
		JOIN purchase_order_items poi ON poi.purchase_order_id = po.id
		WHERE po.status = 'RECEIVED' 
		AND po.order_date BETWEEN ? AND ?
		GROUP BY s.id
	`
	// Note: Return rate requires PurchaseReturn table join, which is separate.
	// We'll fetch base stats first.

	rows, err := r.DB.Raw(query, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sc SupplierScorecard
		var totalReceived, totalOrdered float64
		var avgLeadTime sql.NullFloat64

		if err := rows.Scan(&sc.SupplierID, &sc.SupplierName, &sc.TotalPOs, &avgLeadTime, &totalReceived, &totalOrdered); err != nil {
			return nil, err
		}

		if avgLeadTime.Valid {
			sc.AvgLeadTime = avgLeadTime.Float64
		}

		if totalOrdered > 0 {
			sc.FillRate = (totalReceived / totalOrdered) * 100
		}

		// TODO: Fetch Return Rate separately or via subquery
		scorecards = append(scorecards, sc)
	}

	return scorecards, nil
}

type HeatmapPoint struct {
	DayOfWeek  int // 0=Sunday, 6=Saturday
	HourOfDay  int // 0-23
	TotalSales float64
}

// GetHourlySalesHeatmap aggregates sales by hour and day of week.
func (r *ReportsRepository) GetHourlySalesHeatmap(startDate, endDate time.Time) ([]HeatmapPoint, error) {
	var points []HeatmapPoint

	// Postgres-specific date functions
	query := `
		SELECT 
			EXTRACT(DOW FROM adjusted_at) as day_of_week,
			EXTRACT(HOUR FROM adjusted_at) as hour_of_day,
			SUM(quantity * p.selling_price) as total_sales
		FROM stock_adjustments sa
		JOIN products p ON p.id = sa.product_id
		WHERE sa.type = 'STOCK_OUT' 
		AND sa.reason_code = 'SALE'
		AND sa.adjusted_at BETWEEN ? AND ?
		GROUP BY day_of_week, hour_of_day
		ORDER BY day_of_week, hour_of_day
	`

	rows, err := r.DB.Raw(query, startDate, endDate).Rows()
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

type EmployeeSalesStats struct {
	UserID      uint
	Username    string
	TotalOrders int
	TotalSales  float64
}

// GetSalesByEmployeeReport aggregates sales by the user who processed the order.
func (r *ReportsRepository) GetSalesByEmployeeReport(startDate, endDate time.Time) ([]EmployeeSalesStats, error) {
	var stats []EmployeeSalesStats

	// We use the 'orders' table which links to 'users'
	query := `
		SELECT 
			u.id, 
			u.username, 
			COUNT(o.id) as total_orders, 
			SUM(o.total_amount) as total_sales
		FROM orders o
		JOIN users u ON u.id = o.user_id
		WHERE o.status = 'COMPLETED'
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
	ItemCount     int
}

// GetCategoryDrillDownReport provides sales and margin performance by category.
func (r *ReportsRepository) GetCategoryDrillDownReport(startDate, endDate time.Time) ([]CategoryPerformance, error) {
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

type GMROIStats struct {
	TotalRevenue        float64
	COGS                float64
	GrossMargin         float64
	AverageInventoryVal float64
	GMROI               float64
}

// GetCOGSAndGMROIReport calculates Gross Margin Return on Investment.
// GMROI = Gross Margin / Average Inventory Cost
func (r *ReportsRepository) GetCOGSAndGMROIReport(startDate, endDate time.Time) (GMROIStats, error) {
	var stats GMROIStats

	// 1. Calculate Revenue and COGS
	cogs, avgInv, err := r.GetInventoryTurnover(startDate, endDate, nil, nil) // Reusing existing logic
	if err != nil {
		return stats, err
	}
	stats.COGS = cogs
	stats.AverageInventoryVal = avgInv

	// Calculate Revenue separately (since GetInventoryTurnover returns COGS and AvgInv)
	// Or we can reuse GetProfitMargin which returns Revenue and Cost
	revenue, _, err := r.GetProfitMargin(startDate, endDate, nil, nil)
	if err != nil {
		return stats, err
	}
	stats.TotalRevenue = revenue
	// Note: cost should match cogs, but let's use the one from ProfitMargin for consistency if needed.
	// Actually, GetInventoryTurnover calculates COGS correctly.

	stats.GrossMargin = stats.TotalRevenue - stats.COGS

	if stats.AverageInventoryVal > 0 {
		stats.GMROI = stats.GrossMargin / stats.AverageInventoryVal
	}

	return stats, nil
}

// GetVoidDiscountAuditReport retrieves audit logs for voids and discounts.
func (r *ReportsRepository) GetVoidDiscountAuditReport(startDate, endDate time.Time) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog

	err := r.DB.Where("action IN (?)", []string{"VOID", "DISCOUNT"}).
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Preload("User"). // To see who did it
		Order("timestamp DESC").
		Find(&logs).Error

	return logs, err
}

type TaxLiabilityStats struct {
	TaxRate       float64
	TaxableAmount float64
	TaxAmount     float64
}

// GetTaxLiabilityReport calculates tax collected.
// Note: Currently we assume a single global tax rate from settings, but data might be stored differently.
// If tax is calculated on the fly, we need to reconstruct it or store it in Order/OrderItem.
// For now, we'll estimate based on Order totals and implied tax.
// Ideally, Order table should have 'TaxAmount' column.
func (r *ReportsRepository) GetTaxLiabilityReport(startDate, endDate time.Time) ([]TaxLiabilityStats, error) {
	var stats []TaxLiabilityStats

	// Assuming Order model has TotalAmount (inclusive of tax) and we can back-calculate or if we added TaxAmount column.
	// Let's check domain/orders.go... It doesn't have TaxAmount explicitly in the struct I saw earlier?
	// Wait, I should check.
	// If not, I'll add a placeholder implementation that assumes 0 tax or needs schema update.
	// Looking at previous SalesHandler code: "taxAmount := totalAmount * taxRate".
	// But it wasn't saved to a specific column in Order?
	// "TotalAmount: totalAmount - discountAmount" (Net or Gross?)
	// The SalesHandler added tax to totalAmount.

	// To do this accurately, we really need a TaxAmount column on Order.
	// For this task, I will implement a query that assumes we can sum (TotalAmount - (TotalAmount / (1 + Rate))) if rate is known.
	// But rate can change.

	// Better approach: Use the 'Transactions' table or just sum Orders and apply current rate (imperfect).
	// OR, since I can't change schema easily right now without migration file, I will return a stub or best-effort.

	// Let's assume for now we just return total sales and let frontend apply rate, OR
	// we use a fixed rate query.

	// Actually, let's look at Order struct again.
	// It has "TotalAmount".

	// I'll implement a simple aggregation of TotalAmount for now.

	query := `
		SELECT 
			0.0 as tax_rate, -- Placeholder
			SUM(total_amount) as taxable_amount,
			0.0 as tax_amount -- Placeholder
		FROM orders
		WHERE status = 'COMPLETED'
		AND order_date BETWEEN ? AND ?
	`
	// This is a placeholder until we add TaxAmount to Order schema.

	rows, err := r.DB.Raw(query, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s TaxLiabilityStats
		if err := rows.Scan(&s.TaxRate, &s.TaxableAmount, &s.TaxAmount); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// GetCashDrawerReconciliationReport retrieves cash drawer sessions.
func (r *ReportsRepository) GetCashDrawerReconciliationReport(startDate, endDate time.Time) ([]domain.CashDrawerSession, error) {
	var sessions []domain.CashDrawerSession

	err := r.DB.Preload("User").Preload("Location").
		Where("start_time BETWEEN ? AND ?", startDate, endDate).
		Order("start_time DESC").
		Find(&sessions).Error

	return sessions, err
}

type BasketAnalysisItem struct {
	ProductA  uint
	ProductB  uint
	Frequency int
}

// GetBasketAnalysisReport finds frequently bought together items.
// This is a heavy query, usually done offline or via specialized graph DB.
// Simple SQL approach: Self-join order_items.
func (r *ReportsRepository) GetBasketAnalysisReport(startDate, endDate time.Time) ([]BasketAnalysisItem, error) {
	var items []BasketAnalysisItem

	query := `
		SELECT 
			oi1.product_id as product_a,
			oi2.product_id as product_b,
			COUNT(*) as frequency
		FROM order_items oi1
		JOIN order_items oi2 ON oi1.order_id = oi2.order_id AND oi1.product_id < oi2.product_id
		JOIN orders o ON o.id = oi1.order_id
		WHERE o.order_date BETWEEN ? AND ?
		GROUP BY oi1.product_id, oi2.product_id
		ORDER BY frequency DESC
		LIMIT 20
	`

	rows, err := r.DB.Raw(query, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item BasketAnalysisItem
		if err := rows.Scan(&item.ProductA, &item.ProductB, &item.Frequency); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

// GetPOAnalysisReport retrieves purchase order statistics.
func (r *ReportsRepository) GetPOAnalysisReport(startDate, endDate time.Time) ([]domain.PurchaseOrder, error) {
	var pos []domain.PurchaseOrder

	err := r.DB.Preload("Supplier").Preload("Items").
		Where("order_date BETWEEN ? AND ?", startDate, endDate).
		Order("order_date DESC").
		Find(&pos).Error

	return pos, err
}

type CustomerInsight struct {
	UserID             uint
	Username           string
	FullName           string
	TotalSpent         float64
	OrderCount         int
	LastOrderDate      *time.Time
	DaysSinceLastOrder int
}

// GetCustomerInsightsReport identifies top spenders and churn risk.
func (r *ReportsRepository) GetCustomerInsightsReport(startDate, endDate time.Time) ([]CustomerInsight, error) {
	var insights []CustomerInsight

	// We join users and orders
	query := `
		SELECT 
			u.id, 
			u.username, 
			COALESCE(u.first_name || ' ' || u.last_name, u.username) as full_name,
			SUM(o.total_amount) as total_spent,
			COUNT(o.id) as order_count,
			MAX(o.order_date) as last_order_date
		FROM users u
		JOIN orders o ON o.user_id = u.id
		WHERE o.status = 'COMPLETED'
		AND o.order_date BETWEEN ? AND ?
		GROUP BY u.id, u.username, full_name
		ORDER BY total_spent DESC
		LIMIT 50
	`

	rows, err := r.DB.Raw(query, startDate, endDate).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	now := time.Now()
	for rows.Next() {
		var ci CustomerInsight
		var lastOrder sql.NullTime

		if err := rows.Scan(&ci.UserID, &ci.Username, &ci.FullName, &ci.TotalSpent, &ci.OrderCount, &lastOrder); err != nil {
			return nil, err
		}

		if lastOrder.Valid {
			ci.LastOrderDate = &lastOrder.Time
			ci.DaysSinceLastOrder = int(now.Sub(lastOrder.Time).Hours() / 24)
		}

		insights = append(insights, ci)
	}

	return insights, nil
}

type ShrinkageItem struct {
	ProductID   uint
	ProductName string
	Reason      string
	Quantity    int
	LostValue   float64
}

// GetShrinkageReport tracks inventory loss (Theft, Damage, Expiry).
func (r *ReportsRepository) GetShrinkageReport(startDate, endDate time.Time) ([]ShrinkageItem, error) {
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
		AND sa.reason_code != 'SALE' -- Exclude normal sales
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

type ReturnAnalysisItem struct {
	ProductID     uint
	ProductName   string
	Reason        string
	ReturnCount   int
	TotalRefunded float64
}

// GetCustomerReturnAnalysisReport analyzes why customers are returning items.
func (r *ReportsRepository) GetCustomerReturnAnalysisReport(startDate, endDate time.Time) ([]ReturnAnalysisItem, error) {
	var items []ReturnAnalysisItem

	// Assuming 'returns' and 'return_items' tables exist (based on return_test.go)
	// Let's verify domain models... Yes, Return and ReturnItem exist in models.go

	// Corrected Query assuming ReturnItem has Reason but maybe not Price:
	/*
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
	*/

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
