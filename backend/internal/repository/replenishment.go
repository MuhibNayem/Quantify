package repository

import (
	"inventory/backend/internal/domain"
	"time"
)

func GetSalesDataForForecast(productID uint, days int) ([]domain.StockAdjustment, error) {
	var salesData []domain.StockAdjustment
	periodAgo := time.Now().AddDate(0, 0, -days)

	err := DB.Model(&domain.StockAdjustment{}).
		Where("product_id = ?", productID).
		Where("type = ?", "STOCK_OUT").
		Where("reason_code = ?", "SALE").
		Where("adjusted_at >= ?", periodAgo).
		Order("adjusted_at ASC").
		Find(&salesData).Error

	if err != nil {
		return nil, err
	}

	return salesData, nil
}
