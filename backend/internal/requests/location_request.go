package requests

// LocationCreateRequest represents the request body for creating a new location.
type LocationCreateRequest struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address"`
}

// LocationUpdateRequest represents the request body for updating an existing location.
type LocationUpdateRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}
