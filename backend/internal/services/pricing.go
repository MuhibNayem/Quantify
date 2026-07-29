package services

import (
	"strconv"

	"inventory/backend/internal/domain"
)

func (s *SalesService) applyPromotion(baseUnitPrice float64, product domain.Product, quantity float64, activePromotions []domain.Promotion) (float64, float64) {
	bestPromo := ResolvePromotion(activePromotions, product.ID, product.SubCategoryID, product.CategoryID)

	unitPrice := baseUnitPrice
	discountTotal := 0.0
	if bestPromo != nil {
		if bestPromo.DiscountType == "PERCENTAGE" {
			unitPrice -= baseUnitPrice * (bestPromo.DiscountValue / 100.0)
		} else if bestPromo.DiscountType == "FIXED_AMOUNT" {
			unitPrice -= bestPromo.DiscountValue
		}
		if unitPrice < 0 {
			unitPrice = 0
		}
		discountTotal = (baseUnitPrice - unitPrice) * quantity
	}
	return unitPrice, discountTotal
}

func (s *SalesService) applyTax(amount float64) (postTaxTotal, taxAmount float64, err error) {
	taxRate := 0.0
	if val, err := s.Settings.GetSetting("tax_rate_percentage"); err == nil {
		if v, err := strconv.ParseFloat(val, 64); err == nil {
			taxRate = v / 100.0
		}
	}
	taxAmount = amount * taxRate
	postTaxTotal = amount + taxAmount
	return postTaxTotal, taxAmount, nil
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
