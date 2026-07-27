package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func TestResolvePromotion_Hierarchy(t *testing.T) {
	prodID, subID, catID := uint(1), uint(2), uint(3)
	tests := []struct {
		name   string
		promos []domain.Promotion
		wantID uint
	}{
		{"product beats category", []domain.Promotion{
			{Model: gorm.Model{ID: 1}, CategoryID: &catID, DiscountType: "PERCENTAGE", DiscountValue: 10, Priority: 1},
			{Model: gorm.Model{ID: 2}, ProductID: &prodID, DiscountType: "PERCENTAGE", DiscountValue: 15, Priority: 1},
		}, 2},
		{"subcategory beats category", []domain.Promotion{
			{Model: gorm.Model{ID: 1}, CategoryID: &catID, DiscountType: "PERCENTAGE", DiscountValue: 10},
			{Model: gorm.Model{ID: 2}, SubCategoryID: &subID, DiscountType: "PERCENTAGE", DiscountValue: 12},
		}, 2},
		{"higher priority wins ties", []domain.Promotion{
			{Model: gorm.Model{ID: 1}, ProductID: &prodID, DiscountType: "PERCENTAGE", DiscountValue: 10, Priority: 1},
			{Model: gorm.Model{ID: 2}, ProductID: &prodID, DiscountType: "PERCENTAGE", DiscountValue: 8, Priority: 5},
		}, 2},
		{"empty list returns nil", []domain.Promotion{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolvePromotion(tt.promos, prodID, subID, catID)
			if tt.wantID == 0 {
				assert.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			assert.Equal(t, tt.wantID, got.ID)
		})
	}
}
