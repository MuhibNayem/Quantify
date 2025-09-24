package requests

// CategoryCreateRequest represents the request body for creating a new category.
type CategoryCreateRequest struct {
	Name string `json:"name" binding:"required"`
}

// CategoryUpdateRequest represents the request body for updating an existing category.
type CategoryUpdateRequest struct {
	Name string `json:"name" binding:"required"`
}

// SubCategoryCreateRequest represents the request body for creating a new sub-category.
type SubCategoryCreateRequest struct {
	Name string `json:"name" binding:"required"`
}

// SubCategoryUpdateRequest represents the request body for updating an existing sub-category.
type SubCategoryUpdateRequest struct {
	Name string `json:"name" binding:"required"`
}
