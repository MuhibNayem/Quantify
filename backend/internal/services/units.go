package services

import (
	"fmt"

	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

// resolveProductUnit returns the active ProductUnit for (product, unitID) and
// validates the unit is compatible with the product's base unit (same dimension).
// Returns the ProductUnit whose Factor maps 1 selling unit to Factor base-stock-units.
func (s *SalesService) resolveProductUnit(tx *gorm.DB, product domain.Product, unitID uint) (domain.ProductUnit, error) {
	var pu domain.ProductUnit
	if err := tx.Preload("Unit").Where("product_id = ? AND unit_id = ? AND is_active = ?", product.ID, unitID, true).
		First(&pu).Error; err != nil {
		return pu, fmt.Errorf("product %q is not sold in unit %d", product.Name, unitID)
	}
	// If the product has a base unit, ensure the selling unit shares its dimension.
	if product.BaseUnitID != nil {
		var baseUnit domain.Unit
		if err := tx.First(&baseUnit, *product.BaseUnitID).Error; err == nil {
			if pu.Unit.DimensionID != baseUnit.DimensionID {
				return pu, fmt.Errorf("unit %q is incompatible with product %q (different dimension)", pu.Unit.Name, product.Name)
			}
		}
	}
	if pu.Factor <= 0 {
		pu.Factor = 1
	}
	return pu, nil
}
