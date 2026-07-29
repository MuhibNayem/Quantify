package domain

import "gorm.io/gorm"

// UnitDimension is a physical quantity type. Units within a dimension are
// mutually convertible (you can sell wire in feet, inches, or meters because
// they share the "length" dimension). Countable goods use the "count" dimension.
type UnitDimension struct {
	gorm.Model
	Code       string `gorm:"uniqueIndex;not null" json:"Code"` // length, mass, volume, count, area, time
	Name       string `gorm:"not null" json:"Name"`
	BaseUnitID *uint  `json:"BaseUnitID"` // canonical unit for this dimension
}

// Unit is a unit of measure within a dimension.
//
// Factor expresses how many of the dimension's BASE units equal exactly 1 of
// this unit. With length base = foot:
//
//	foot.Factor  = 1.0
//	inch.Factor  = 1/12      (1 inch  = 0.08333 ft)
//	meter.Factor = 3.28084   (1 meter = 3.28084 ft)
//	yard.Factor  = 3.0
//
// Conversion between two units of the same dimension is therefore:
//
//	qtyInB = qtyInA * (A.Factor / B.Factor)
//
// Countable units work the same way (each=1, pair=2, dozen=12).
type Unit struct {
	gorm.Model
	Code        string  `gorm:"uniqueIndex;not null" json:"Code"` // ft, in, m, kg, lb, each, dozen
	Name        string  `gorm:"not null" json:"Name"`             // Foot, Inch, Meter
	Symbol      string  `json:"Symbol"`                           // ft, in, m
	DimensionID uint    `gorm:"not null;index" json:"DimensionID"`
	Dimension   UnitDimension
	Factor      float64 `gorm:"not null;default:1" json:"Factor"` // relative to dimension base unit
	IsBase      bool    `gorm:"default:false" json:"IsBase"`
}

// ProductUnit defines a unit a specific product can be transacted in, together
// with the conversion to the product's base STOCK unit and an optional
// per-unit price.
//
// Example — copper wire stocked by the foot (BaseUnit = foot):
//
//	coil: Factor = 100   → 1 coil = 100 ft,  Price = 45.00  (buy/sell whole coils)
//	inch: Factor = 1/12  → 1 inch = 0.0833 ft, Price = 0.05 (sell by the inch)
//	foot: Factor = 1     → 1 foot = 1 ft,    Price = 0.50  (sell by the foot)
//
// Selling 14 inches deducts 14 * (1/12) = 1.1667 ft from stock automatically.
type ProductUnit struct {
	gorm.Model
	ProductID     uint    `gorm:"not null;uniqueIndex:idx_product_unit" json:"ProductID"`
	Product       Product `json:"-"`
	UnitID        uint    `gorm:"not null;uniqueIndex:idx_product_unit" json:"UnitID"`
	Unit          Unit    `json:"Unit"`
	Factor        float64 `gorm:"not null;default:1" json:"Factor"` // 1 of this unit = Factor base-stock-units
	Price         float64 `gorm:"default:0" json:"Price"`           // selling price per this unit (0 → derive)
	PurchasePrice float64 `gorm:"default:0" json:"PurchasePrice"`   // purchase price per this unit (0 → derive)
	Barcode       string  `gorm:"index" json:"Barcode"`             // unit-specific barcode (a coil ≠ a foot)
	IsActive      bool    `gorm:"default:true" json:"IsActive"`
}
