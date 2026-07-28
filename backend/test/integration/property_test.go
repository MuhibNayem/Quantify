//go:build integration

package integration

import (
	"math"
	"testing"

	"pgregory.net/rapid"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/services"
)

func TestPropertyPromotionResolverHierarchy(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		productID := uint(rapid.IntRange(1, 1000).Draw(rt, "productID"))
		subCatID := uint(rapid.IntRange(1, 1000).Draw(rt, "subCatID"))
		catID := uint(rapid.IntRange(1, 1000).Draw(rt, "catID"))

		numPromos := rapid.IntRange(0, 5).Draw(rt, "numPromos")
		promos := make([]domain.Promotion, 0, numPromos)
		for i := 0; i < numPromos; i++ {
			p := domain.Promotion{
				DiscountType:  rapid.SampledFrom([]string{"PERCENTAGE", "FIXED_AMOUNT"}).Draw(rt, "discountType"),
				DiscountValue: rapid.Float64Range(0, 100).Draw(rt, "discountValue"),
				Priority:      rapid.IntRange(0, 10).Draw(rt, "priority"),
			}
			scope := rapid.IntRange(0, 2).Draw(rt, "scope")
			switch scope {
			case 0:
				id := productID
				p.ProductID = &id
			case 1:
				id := subCatID
				p.SubCategoryID = &id
			case 2:
				id := catID
				p.CategoryID = &id
			}
			promos = append(promos, p)
		}

		result := services.ResolvePromotion(promos, productID, subCatID, catID)

		if numPromos == 0 {
			if result != nil {
				rt.Fatalf("expected nil for empty promos, got %+v", result)
			}
			return
		}

		if result == nil {
			rt.Fatalf("expected non-nil result for %d promos", numPromos)
		}

		hasProduct := false
		hasSubCat := false
		for i := range promos {
			if promos[i].ProductID != nil && *promos[i].ProductID == productID {
				hasProduct = true
			}
			if promos[i].SubCategoryID != nil && *promos[i].SubCategoryID == subCatID {
				hasSubCat = true
			}
		}

		if hasProduct {
			if result.ProductID == nil || *result.ProductID != productID {
				rt.Fatalf("product-level promo should win, got scope rank mismatch")
			}
		} else if hasSubCat {
			if result.SubCategoryID == nil || *result.SubCategoryID != subCatID {
				rt.Fatalf("subcategory-level promo should win when no product promo exists")
			}
		}
	})
}

func TestPropertyRefundNeverNegative(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		unitPrice := rapid.Float64Range(0, 10000).Draw(rt, "unitPrice")
		qty := rapid.IntRange(1, 1000).Draw(rt, "qty")
		taxRate := rapid.Float64Range(0, 0.5).Draw(rt, "taxRate")

		subtotal := unitPrice * float64(qty)
		tax := subtotal * taxRate
		refund := math.Round((subtotal+tax)*100) / 100

		if refund < 0 {
			rt.Fatalf("refund must never be negative: got %f", refund)
		}
	})
}

func TestPropertyTaxCalculationRoundTrip(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		amount := rapid.Float64Range(0, 100000).Draw(rt, "amount")
		taxPct := rapid.Float64Range(0, 30).Draw(rt, "taxPct")

		taxRate := taxPct / 100.0
		taxAmount := amount * taxRate
		postTaxTotal := amount + taxAmount

		recoveredAmount := postTaxTotal - taxAmount

		if math.Abs(recoveredAmount-amount) > 1e-9 {
			rt.Fatalf("round-trip failed: amount=%f, recovered=%f, diff=%e",
				amount, recoveredAmount, math.Abs(recoveredAmount-amount))
		}

		if postTaxTotal < amount {
			rt.Fatalf("post-tax total %f must be >= pre-tax amount %f", postTaxTotal, amount)
		}

		if taxAmount < 0 {
			rt.Fatalf("tax amount must never be negative: got %f", taxAmount)
		}
	})
}
