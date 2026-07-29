package repository

import (
	"database/sql"
	"inventory/backend/internal/domain"
	"time"
)

func (r *reportsRepository) GetInventoryTurnover(startDate, endDate time.Time, categoryID, locationID *uint) (float64, float64, error) {
	var costOfGoodsSold sql.NullFloat64

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

func (r *reportsRepository) getInventoryValueAt(date time.Time, categoryID, locationID *uint) (float64, error) {
	var totalValue sql.NullFloat64

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

type StockAgingItem struct {
	ProductID   uint
	ProductName string
	SKU         string
	AgeDays     int
	Quantity    float64
	Value       float64
}

func (r *reportsRepository) GetStockAgingReport() (map[string][]StockAgingItem, error) {
	var batches []domain.Batch
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
			Value:       b.Quantity * b.Product.PurchasePrice,
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
	CurrentStock  float64
	LastSaleDate  *time.Time
	DaysSinceSale int
	Value         float64
}

func (r *reportsRepository) GetDeadStockReport(daysThreshold int) ([]DeadStockItem, error) {
	var items []DeadStockItem
	thresholdDate := time.Now().AddDate(0, 0, -daysThreshold)

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
		HAVING COALESCE(SUM(b.quantity), 0) > 0 
		AND (MAX(sa.adjusted_at) < ? OR MAX(sa.adjusted_at) IS NULL)
		AND MAX(b.created_at) < ?
		ORDER BY last_sale_date ASC NULLS FIRST
	`

	rows, err := r.DB.Raw(query, thresholdDate, thresholdDate).Rows()
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
			item.DaysSinceSale = -1
		}
		item.Value = item.CurrentStock * price
		items = append(items, item)
	}

	return items, nil
}

type GMROIStats struct {
	TotalRevenue        float64
	COGS                float64
	GrossMargin         float64
	AverageInventoryVal float64
	GMROI               float64
}

func (r *reportsRepository) GetCOGSAndGMROIReport(startDate, endDate time.Time) (GMROIStats, error) {
	var stats GMROIStats

	cogs, avgInv, err := r.GetInventoryTurnover(startDate, endDate, nil, nil)
	if err != nil {
		return stats, err
	}
	stats.COGS = cogs
	stats.AverageInventoryVal = avgInv

	revenue, _, err := r.GetProfitMargin(startDate, endDate, nil, nil)
	if err != nil {
		return stats, err
	}
	stats.TotalRevenue = revenue

	stats.GrossMargin = stats.TotalRevenue - stats.COGS

	if stats.AverageInventoryVal > 0 {
		stats.GMROI = stats.GrossMargin / stats.AverageInventoryVal
	}

	return stats, nil
}
