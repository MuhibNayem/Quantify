package units

import (
	"math"
	"testing"
)

func almost(a, b float64) bool { return math.Abs(a-b) < 1e-4 }

func TestToBase_InchesToFeet(t *testing.T) {
	inchFactor := 1.0 / 12.0
	// 14 inches = 1.1667 ft
	if got := ToBase(14, inchFactor); !almost(got, 1.1667) {
		t.Fatalf("14 in → ft = %v, want 1.1667", got)
	}
	// 1.2 feet (factor 1) = 1.2
	if got := ToBase(1.2, 1.0); !almost(got, 1.2) {
		t.Fatalf("1.2 ft = %v, want 1.2", got)
	}
}

func TestStockDeduction_Coil(t *testing.T) {
	// 1 coil = 100 ft of wire
	if got := StockDeduction(1, 100); !almost(got, 100) {
		t.Fatalf("1 coil = %v ft, want 100", got)
	}
	// 2.5 coils = 250 ft
	if got := StockDeduction(2.5, 100); !almost(got, 250) {
		t.Fatalf("2.5 coils = %v ft, want 250", got)
	}
}

func TestConvert_YardsToInches(t *testing.T) {
	yardFactor := 3.0
	inchFactor := 1.0 / 12.0
	// 2 yards = 72 inches
	if got := Convert(2, yardFactor, inchFactor); !almost(got, 72) {
		t.Fatalf("2 yd → in = %v, want 72", got)
	}
}

func TestConvert_MetersToFeet(t *testing.T) {
	meterFactor := 3.28084
	footFactor := 1.0
	// 1 meter = 3.28084 ft
	if got := Convert(1, meterFactor, footFactor); !almost(got, 3.28084) {
		t.Fatalf("1 m → ft = %v, want 3.28084", got)
	}
}

func TestFromBase_FeetToInches(t *testing.T) {
	inchFactor := 1.0 / 12.0
	if got := FromBase(5, inchFactor); !almost(got, 60) {
		t.Fatalf("5 ft → in = %v, want 60", got)
	}
}

func TestEffectivePrice(t *testing.T) {
	// Explicit unit price wins
	if got := EffectivePrice(45.0, 0.5, 100); !almost(got, 45.0) {
		t.Fatalf("explicit price = %v, want 45", got)
	}
	// Derived: $0.50/ft × 100 ft/coil = $50/coil
	if got := EffectivePrice(0, 0.5, 100); !almost(got, 50.0) {
		t.Fatalf("derived coil price = %v, want 50", got)
	}
}

func TestRoundQty(t *testing.T) {
	if got := RoundQty(1.166666666); !almost(got, 1.1667) {
		t.Fatalf("RoundQty = %v, want 1.1667", got)
	}
}

func TestConvert_ZeroToFactor(t *testing.T) {
	if got := Convert(5, 1.0, 0); got != 0 {
		t.Fatalf("divide-by-zero guard failed: %v", got)
	}
}
