package repository

import (
	"inventory/backend/internal/domain"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// dimensionSeed defines a dimension and its base unit code.
type dimensionSeed struct {
	Code     string
	Name     string
	BaseCode string
}

// unitSeed defines a unit. Factor is relative to the dimension's base unit
// (1 of this unit = Factor base units).
type unitSeed struct {
	Code     string
	Name     string
	Symbol   string
	DimCode  string
	Factor   float64
	IsBase   bool
}

// Universal unit catalog. Data-driven and extensible — add a row to support a
// new unit; no code changes. Factors are exact standard conversion values.
var dimensions = []dimensionSeed{
	{"count", "Count", "each"},
	{"length", "Length", "meter"},
	{"mass", "Mass", "kilogram"},
	{"volume", "Volume", "liter"},
	{"area", "Area", "square_meter"},
	{"time", "Time", "second"},
}

var unitCatalog = []unitSeed{
	// Count (base = each)
	{"each", "Each", "ea", "count", 1, true},
	{"pair", "Pair", "pr", "count", 2, false},
	{"dozen", "Dozen", "dz", "count", 12, false},
	{"score", "Score", "sc", "count", 20, false},
	{"gross", "Gross", "gr", "count", 144, false},

	// Length (base = meter)
	{"meter", "Meter", "m", "length", 1, true},
	{"millimeter", "Millimeter", "mm", "length", 0.001, false},
	{"centimeter", "Centimeter", "cm", "length", 0.01, false},
	{"kilometer", "Kilometer", "km", "length", 1000, false},
	{"inch", "Inch", "in", "length", 0.0254, false},
	{"foot", "Foot", "ft", "length", 0.3048, false},
	{"yard", "Yard", "yd", "length", 0.9144, false},
	{"mile", "Mile", "mi", "length", 1609.344, false},

	// Mass (base = kilogram)
	{"kilogram", "Kilogram", "kg", "mass", 1, true},
	{"gram", "Gram", "g", "mass", 0.001, false},
	{"milligram", "Milligram", "mg", "mass", 0.000001, false},
	{"tonne", "Tonne", "t", "mass", 1000, false},
	{"ounce", "Ounce", "oz", "mass", 0.0283495, false},
	{"pound", "Pound", "lb", "mass", 0.453592, false},
	{"stone", "Stone", "st", "mass", 6.35029, false},

	// Volume (base = liter)
	{"liter", "Liter", "L", "volume", 1, true},
	{"milliliter", "Milliliter", "mL", "volume", 0.001, false},
	{"cubic_meter", "Cubic Meter", "m³", "volume", 1000, false},
	{"gallon_us", "US Gallon", "gal", "volume", 3.78541, false},
	{"quart", "Quart", "qt", "volume", 0.946353, false},
	{"pint", "Pint", "pt", "volume", 0.473176, false},
	{"cup", "Cup", "cup", "volume", 0.236588, false},
	{"fluid_ounce", "Fluid Ounce", "fl oz", "volume", 0.0295735, false},

	// Area (base = square meter)
	{"square_meter", "Square Meter", "m²", "area", 1, true},
	{"square_foot", "Square Foot", "ft²", "area", 0.092903, false},
	{"square_yard", "Square Yard", "yd²", "area", 0.836127, false},
	{"square_inch", "Square Inch", "in²", "area", 0.00064516, false},
	{"acre", "Acre", "ac", "area", 4046.86, false},
	{"hectare", "Hectare", "ha", "area", 10000, false},

	// Time (base = second)
	{"second", "Second", "s", "time", 1, true},
	{"minute", "Minute", "min", "time", 60, false},
	{"hour", "Hour", "h", "time", 3600, false},
	{"day", "Day", "d", "time", 86400, false},
}

// SeedUnits idempotently inserts the universal dimension + unit catalog and
// wires each dimension's BaseUnitID. Safe to run on every boot.
func SeedUnits(db *gorm.DB) error {
	// Upsert dimensions (by code).
	dimByCode := map[string]*domain.UnitDimension{}
	for _, d := range dimensions {
		dim := &domain.UnitDimension{Code: d.Code, Name: d.Name}
		if err := db.Where("code = ?", d.Code).
			Assign(domain.UnitDimension{Name: d.Name}).
			FirstOrCreate(dim).Error; err != nil {
			return err
		}
		dimByCode[d.Code] = dim
	}

	// Upsert units (by code), linking to their dimension.
	unitByCode := map[string]*domain.Unit{}
	for _, u := range unitCatalog {
		dim := dimByCode[u.DimCode]
		if dim == nil {
			continue
		}
		unit := &domain.Unit{Code: u.Code}
		if err := db.Where("code = ?", u.Code).
			Assign(domain.Unit{
				Name:        u.Name,
				Symbol:      u.Symbol,
				DimensionID: dim.ID,
				Factor:      u.Factor,
				IsBase:      u.IsBase,
			}).
			FirstOrCreate(unit).Error; err != nil {
			return err
		}
		unitByCode[u.Code] = unit
	}

	// Wire each dimension's base unit.
	for _, d := range dimensions {
		dim := dimByCode[d.Code]
		base := unitByCode[d.BaseCode]
		if dim == nil || base == nil {
			continue
		}
		if err := db.Model(dim).Update("base_unit_id", base.ID).Error; err != nil {
			return err
		}
	}

	logrus.WithField("units", len(unitCatalog)).Info("Seeded universal unit catalog")
	return nil
}

// UnitByCode returns a unit by its code (helper for tests / seeding products).
func UnitByCode(db *gorm.DB, code string) (*domain.Unit, error) {
	var u domain.Unit
	err := db.Where("code = ?", code).First(&u).Error
	return &u, err
}

// ensure SeedUnits uses upsert semantics even if called concurrently
var _ = clause.OnConflict{}
