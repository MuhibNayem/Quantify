package requests

import "time"

type PromotionCreateRequest struct {
	Name          string    `json:"name" binding:"required"`
	Description   string    `json:"description"`
	DiscountType  string    `json:"discountType" binding:"required,oneof=PERCENTAGE FIXED_AMOUNT"`
	DiscountValue float64   `json:"discountValue" binding:"required,gt=0"`
	StartDate     time.Time `json:"startDate" binding:"required"`
	EndDate       time.Time `json:"endDate" binding:"required,gtfield=StartDate"`
	Priority      int       `json:"priority"`
	ProductID     *uint     `json:"productId"`
	CategoryID    *uint     `json:"categoryId"`
	SubCategoryID *uint     `json:"subCategoryId"`
}

type PromotionUpdateRequest struct {
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	DiscountType  string    `json:"discountType" binding:"omitempty,oneof=PERCENTAGE FIXED_AMOUNT"`
	DiscountValue float64   `json:"discountValue" binding:"omitempty,gt=0"`
	StartDate     time.Time `json:"startDate"`
	EndDate       time.Time `json:"endDate" binding:"omitempty,gtfield=StartDate"`
	IsActive      *bool     `json:"isActive"`
	Priority      int       `json:"priority"`
}
