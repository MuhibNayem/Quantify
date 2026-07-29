// Package units implements a generic, dynamic unit-of-measure conversion
// engine. It is dimension-aware: any two units sharing a dimension (length,
// mass, volume, count…) are convertible through their factors relative to the
// dimension's base unit. Nothing is hardcoded — units and factors are data.
package units

import "math"

// QtyPrecision is the number of decimal places quantities are rounded to.
// Measured goods (wire, fabric, fuel) are inherently fractional; 4dp gives
// sub-millimeter / sub-milligram resolution while avoiding float noise.
const QtyPrecision = 4

var qtyScale = math.Pow10(QtyPrecision)

// RoundQty normalizes a quantity to QtyPrecision decimal places.
func RoundQty(q float64) float64 {
	return math.Round(q*qtyScale) / qtyScale
}

// ToBase converts a quantity expressed in a unit (whose Factor is relative to
// the dimension base) into the base unit.
//
//	Sell 14 inches of wire (inch.Factor = 1/12): ToBase(14, 1/12) = 1.1667 ft
func ToBase(qty, unitFactor float64) float64 {
	return RoundQty(qty * unitFactor)
}

// FromBase converts a base-unit quantity into a target unit.
//
//	5 ft expressed in inches (inch.Factor = 1/12): FromBase(5, 1/12) = 60 in
func FromBase(qty, unitFactor float64) float64 {
	if unitFactor == 0 {
		return 0
	}
	return RoundQty(qty / unitFactor)
}

// Convert converts a quantity directly between two units of the same dimension
// using their base-relative factors: qtyInTo = qtyInFrom * (fromFactor / toFactor).
//
//	2 yards to inches: Convert(2, yardFactor=3, inchFactor=1/12) = 72 in
func Convert(qty, fromFactor, toFactor float64) float64 {
	if toFactor == 0 {
		return 0
	}
	return RoundQty(qty * fromFactor / toFactor)
}

// StockDeduction computes how much base-stock to remove when selling `qty` of a
// product in a given selling unit. productUnitFactor is the ProductUnit.Factor
// (1 selling unit = Factor base-stock-units).
//
//	Sell 1 coil (coil.Factor = 100 ft): StockDeduction(1, 100) = 100 ft
//	Sell 1.2 ft (foot.Factor = 1):      StockDeduction(1.2, 1) = 1.2 ft
func StockDeduction(qty, productUnitFactor float64) float64 {
	if productUnitFactor == 0 {
		productUnitFactor = 1
	}
	return RoundQty(qty * productUnitFactor)
}

// EffectivePrice returns the price for one selling unit. If the ProductUnit has
// an explicit price (> 0) it wins; otherwise the price is derived from the
// per-base-unit price scaled by the unit factor (price per coil = price/ft × 100).
func EffectivePrice(unitPrice, baseUnitPrice, productUnitFactor float64) float64 {
	if unitPrice > 0 {
		return unitPrice
	}
	return RoundQty(baseUnitPrice * productUnitFactor)
}
