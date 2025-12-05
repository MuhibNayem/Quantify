package repository

import (
	"inventory/backend/internal/domain"
	"time"

	"gorm.io/gorm"
)

type SupplierRepository struct {
	db *gorm.DB
}

func NewSupplierRepository(db *gorm.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

func (r *SupplierRepository) CreateSupplier(supplier *domain.Supplier) error {

	return r.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Create(supplier).Error; err != nil {

			return err

		}

		return UpdateSearchIndex(tx, supplier)

	})

}

func (r *SupplierRepository) UpdateSupplier(supplier *domain.Supplier, updates map[string]interface{}) error {

	return r.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(supplier).Updates(updates).Error; err != nil {

			return err

		}

		if err := tx.First(supplier, supplier.ID).Error; err != nil {

			return err

		}

		return UpdateSearchIndex(tx, supplier)

	})

}

func (r *SupplierRepository) DeleteSupplier(supplier *domain.Supplier) error {

	return r.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Delete(supplier).Error; err != nil {

			return err

		}

		return DeleteFromSearchIndex(tx, supplier.GetEntityType(), supplier.GetID())

	})

}

func (r *SupplierRepository) GetAll() ([]domain.Supplier, error) {
	var suppliers []domain.Supplier
	if err := r.db.Find(&suppliers).Error; err != nil {
		return nil, err
	}
	return suppliers, nil
}

func (r *SupplierRepository) GetByNames(names []string, suppliers *[]domain.Supplier) error {
	return r.db.Where("name IN ?", names).Find(suppliers).Error
}

func (r *SupplierRepository) GetSupplierByName(name string) (*domain.Supplier, error) {

	var supplier domain.Supplier

	if err := r.db.Where("name = ?", name).First(&supplier).Error; err != nil {

		return nil, err

	}

	return &supplier, nil

}

func (r *SupplierRepository) GetSupplierPerformance(supplierID uint) (float64, float64, error) {

	var purchaseOrders []domain.PurchaseOrder

	err := r.db.Where("supplier_id = ? AND status = ?", supplierID, "RECEIVED").Find(&purchaseOrders).Error

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
