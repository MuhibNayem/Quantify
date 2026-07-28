package repository

import (
	"database/sql"
	"inventory/backend/internal/domain"
	"strconv"
	"time"
)

func (r *reportsRepository) GetProfitMargin(startDate, endDate time.Time, categoryID, locationID *uint) (float64, float64, error) {
	var totalRevenue sql.NullFloat64
	var totalCost sql.NullFloat64

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

type SupplierScorecard struct {
	SupplierID   uint
	SupplierName string
	TotalPOs     int
	AvgLeadTime  float64
	FillRate     float64
	ReturnRate   float64
}

func (r *reportsRepository) GetSupplierPerformanceReport(startDate, endDate time.Time) ([]SupplierScorecard, error) {
	scorecards := []SupplierScorecard{}

	query := `
		SELECT 
			s.id, 
			s.name, 
			COUNT(DISTINCT po.id) as total_pos,
			AVG(EXTRACT(EPOCH FROM (po.updated_at - po.order_date))/86400) as avg_lead_time,
			SUM(poi.received_quantity) as total_received,
			SUM(poi.ordered_quantity) as total_ordered
		FROM suppliers s
		JOIN purchase_orders po ON po.supplier_id = s.id
		JOIN purchase_order_items poi ON poi.purchase_order_id = po.id
		WHERE po.status = 'RECEIVED' 
		AND po.order_date BETWEEN ? AND ?
		GROUP BY s.id
	`

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

		scorecards = append(scorecards, sc)
	}

	return scorecards, nil
}

type TaxLiabilityStats struct {
	TaxRate       float64
	TaxableAmount float64
	TaxAmount     float64
}

func (r *reportsRepository) GetTaxLiabilityReport(startDate, endDate time.Time) ([]TaxLiabilityStats, error) {
	var stats []TaxLiabilityStats

	var taxRateSetting domain.SystemSetting
	var taxRate float64 = 0.0

	if err := r.DB.Where("key = ?", "tax_rate_percentage").First(&taxRateSetting).Error; err == nil {
		if v, err := strconv.ParseFloat(taxRateSetting.Value, 64); err == nil {
			taxRate = v / 100.0
		}
	}

	var totalSales sql.NullFloat64
	query := `
		SELECT COALESCE(SUM(total_amount), 0)
		FROM orders
		WHERE status = 'COMPLETED'
		AND order_date BETWEEN ? AND ?
	`
	if err := r.DB.Raw(query, startDate, endDate).Scan(&totalSales).Error; err != nil {
		return nil, err
	}

	totalAmount := 0.0
	if totalSales.Valid {
		totalAmount = totalSales.Float64
	}

	taxableAmount := totalAmount / (1 + taxRate)
	taxAmount := totalAmount - taxableAmount

	stats = append(stats, TaxLiabilityStats{
		TaxRate:       taxRate * 100,
		TaxableAmount: taxableAmount,
		TaxAmount:     taxAmount,
	})

	return stats, nil
}

func (r *reportsRepository) GetCashDrawerReconciliationReport(startDate, endDate time.Time) ([]domain.CashDrawerSession, error) {
	var sessions []domain.CashDrawerSession

	err := r.DB.Preload("User").Preload("Location").
		Where("start_time BETWEEN ? AND ?", startDate, endDate).
		Order("start_time DESC").
		Find(&sessions).Error

	return sessions, err
}

func (r *reportsRepository) GetVoidDiscountAuditReport(startDate, endDate time.Time) ([]domain.AuditLog, error) {
	var logs []domain.AuditLog

	err := r.DB.Where("action IN (?)", []string{"VOID", "DISCOUNT"}).
		Where("timestamp BETWEEN ? AND ?", startDate, endDate).
		Preload("User").
		Order("timestamp DESC").
		Find(&logs).Error

	return logs, err
}

func (r *reportsRepository) GetPOAnalysisReport(startDate, endDate time.Time) ([]domain.PurchaseOrder, error) {
	var pos []domain.PurchaseOrder

	err := r.DB.Preload("Supplier").Preload("Items").
		Where("order_date BETWEEN ? AND ?", startDate, endDate).
		Order("order_date DESC").
		Find(&pos).Error

	return pos, err
}
