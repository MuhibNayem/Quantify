package repository

import (
	"inventory/backend/internal/domain"
	"time"
)

func GetSupplierPerformance(supplierID uint) (float64, float64, error) {
	var purchaseOrders []domain.PurchaseOrder

	err := DB.Where("supplier_id = ? AND status = ?", supplierID, "RECEIVED").Find(&purchaseOrders).Error
	if err != nil {
		return 0, 0, err
	}

	if len(purchaseOrders) == 0 {
		return 0, 0, nil
	}

	var totalLeadTime time.Duration
	var onTimeDeliveries int

	for _, po := range purchaseOrders {
		if po.ActualDeliveryDate != nil && po.ExpectedDeliveryDate != nil {
			leadTime := po.ActualDeliveryDate.Sub(po.OrderDate)
			totalLeadTime += leadTime

			if po.ActualDeliveryDate.Before(*po.ExpectedDeliveryDate) || po.ActualDeliveryDate.Equal(*po.ExpectedDeliveryDate) {
				onTimeDeliveries++
			}
		}
	}

	averageLeadTime := totalLeadTime.Hours() / 24 / float64(len(purchaseOrders))
	onTimeDeliveryRate := float64(onTimeDeliveries) / float64(len(purchaseOrders))

	return averageLeadTime, onTimeDeliveryRate, nil
}
