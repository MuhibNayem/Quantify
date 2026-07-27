package services

import "inventory/backend/internal/domain"

// ResolvePromotion picks the most specific applicable promotion.
// Hierarchy: product > subcategory > category.
// Ties: higher Priority, then higher DiscountValue.
func ResolvePromotion(promos []domain.Promotion, productID, subCategoryID, categoryID uint) *domain.Promotion {
	var best *domain.Promotion
	for i := range promos {
		p := &promos[i]
		if !applies(p, productID, subCategoryID, categoryID) {
			continue
		}
		if best == nil || scopeRank(p) > scopeRank(best) || (scopeRank(p) == scopeRank(best) && higherRank(p, best)) {
			best = p
		}
	}
	return best
}

func applies(p *domain.Promotion, productID, subID, catID uint) bool {
	if p.ProductID != nil && *p.ProductID == productID {
		return true
	}
	if p.SubCategoryID != nil && *p.SubCategoryID == subID {
		return true
	}
	if p.CategoryID != nil && *p.CategoryID == catID {
		return true
	}
	return false
}

func scopeRank(p *domain.Promotion) int {
	if p.ProductID != nil {
		return 3
	}
	if p.SubCategoryID != nil {
		return 2
	}
	return 1
}

func higherRank(a, b *domain.Promotion) bool {
	if a.Priority != b.Priority {
		return a.Priority > b.Priority
	}
	return a.DiscountValue > b.DiscountValue
}
