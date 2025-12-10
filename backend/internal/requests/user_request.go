package requests

// UserRegisterRequest represents the request body for user registration.
type UserRegisterRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Password    string `json:"password" binding:"required,min=6"`
	Role        string `json:"role" binding:"required"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Email       string `json:"email,omitempty" binding:"omitempty,email"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Address     string `json:"address,omitempty"`
}

// UserLoginRequest represents the request body for user login.
type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest represents the request body for refreshing an access token.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// UserUpdateRequest represents the request body for updating user information.
type UserUpdateRequest struct {
	Username    string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	Password    string `json:"password,omitempty" binding:"omitempty,min=6"`
	Role        string `json:"role,omitempty" binding:"omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Email       string `json:"email,omitempty" binding:"omitempty,email"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	Address     string `json:"address,omitempty"`
}

// CreateCustomerRequest represents the request body for creating a new customer.
type CreateCustomerRequest struct {
	Username    string `json:"username" binding:"required,min=3,max=50"`
	Password    string `json:"password" binding:"required,min=6"`
	FirstName   string `json:"firstName" binding:"required"`
	LastName    string `json:"lastName" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
}

// UpdateCustomerRequest represents the request body for updating a customer.
type UpdateCustomerRequest struct {
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Email       string `json:"email,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
}
