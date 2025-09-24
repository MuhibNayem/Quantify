package requests

// SupplierCreateRequest represents the request body for creating a new supplier.
type SupplierCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	ContactPerson string `json:"contactPerson"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
}

// SupplierUpdateRequest represents the request body for updating an existing supplier.
type SupplierUpdateRequest struct {
	Name        string `json:"name"`
	ContactPerson string `json:"contactPerson"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	Address     string `json:"address"`
}
