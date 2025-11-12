package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"inventory/backend/internal/auth"
	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
)

type UserHandler struct {
	userRepo *repository.UserRepository
	db       *gorm.DB
}

func NewUserHandler(userRepo *repository.UserRepository, db *gorm.DB) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		db:       db,
	}
}

// ListUsers godoc
// @Summary List users with optional status and search filters
// @Description Retrieves all users, optionally filtered by status (approved/pending) and search query (username or ID)
// @Tags users
// @Accept json
// @Produce json
// @Param status query string false "Filter by user status (approved, pending)"
// @Param q query string false "Search by username or ID"
// @Security ApiKeyAuth
// @Success 200 {array} domain.User
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var users []domain.User
	db := h.db

	switch strings.ToLower(c.Query("status")) {
	case "approved":
		db = db.Where("is_active = ?", true)
	case "pending":
		db = db.Where("is_active = ?", false)
	}

	if q := c.Query("q"); q != "" {
		pattern := fmt.Sprintf("%%%s%%", q)
		db = db.Where("username ILIKE ? OR CAST(id AS TEXT) ILIKE ?", pattern, pattern)
	}

	if err := db.Order("id ASC").Find(&users).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to list users", http.StatusInternalServerError, err))
		return
	}

	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, users)
}

// RegisterUser godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Param user body requests.UserRegisterRequest true "User registration request"
// @Success 201 {object} domain.User
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/register [post]
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var req requests.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	// Check if any user already exists
	var userCount int64
	if err := h.db.Model(&domain.User{}).Count(&userCount).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to check for existing users", http.StatusInternalServerError, err))
		return
	}

	var user domain.User
	if userCount == 0 {
		// This is the first user, they must be an Admin and will be active
		if req.Role != "Admin" {
			c.Error(appErrors.NewAppError("The first user must be an Admin", http.StatusBadRequest, nil))
			return
		}
		user = domain.User{
			Username: req.Username,
			Role:     req.Role,
			IsActive: true, // First user is active by default
		}
	} else {
		// Subsequent users are not active by default
		user = domain.User{
			Username: req.Username,
			Role:     req.Role,
			IsActive: false, // Subsequent users are inactive by default
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to hash password", http.StatusInternalServerError, err))
		return
	}
	user.Password = string(hashedPassword)

	if err := h.userRepo.CreateUser(&user); err != nil {
		c.Error(appErrors.NewAppError("Failed to create user", http.StatusInternalServerError, err))
		return
	}

	user.Password = ""
	c.JSON(http.StatusCreated, user)
}

// LoginUser godoc
// @Summary Log in a user
// @Description Authenticate user and return a JWT token
// @Tags users
// @Accept json
// @Produce json
// @Param user body requests.UserLoginRequest true "User login request"
// @Success 200 {object} map[string]interface{} "Login successful, returns JWT token"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized: Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/login [post]
func (h *UserHandler) LoginUser(c *gin.Context) {
	var req requests.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var user domain.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Invalid credentials", http.StatusUnauthorized, nil))
			return
		}
		c.Error(appErrors.NewAppError("Database error", http.StatusInternalServerError, err))
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.Error(appErrors.NewAppError("Account not active. Please contact an administrator.", http.StatusUnauthorized, nil))
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.Error(appErrors.NewAppError("Invalid credentials", http.StatusUnauthorized, nil))
		return
	}

	accessToken, refreshToken, err := auth.GenerateTokens(user.ID, user.Role)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to generate tokens", http.StatusInternalServerError, err))
		return
	}

	// Store tokens in Redis
	accessTokenExpiresAt := time.Now().Add(8 * time.Hour)
	if err := repository.SetCache("access_token:"+accessToken, user.ID, accessTokenExpiresAt.Sub(time.Now())); err != nil {
		c.Error(appErrors.NewAppError("Failed to store access token", http.StatusInternalServerError, err))
		return
	}

	refreshTokenExpiresAt := time.Now().Add(10 * 365 * 24 * time.Hour) // 10 years
	if err := repository.SetCache("refresh_token:"+refreshToken, user.ID, refreshTokenExpiresAt.Sub(time.Now())); err != nil {
		c.Error(appErrors.NewAppError("Failed to store refresh token", http.StatusInternalServerError, err))
		return
	}

	csrfToken, err := auth.GenerateCSRFToken(user.ID)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to generate CSRF token", http.StatusInternalServerError, err))
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"csrfToken":    csrfToken,
		"user":         user,
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags users
// @Accept json
// @Produce json
// @Param token body requests.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} map[string]interface{} "New access token"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized: Invalid refresh token"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/refresh-token [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req requests.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	// Check if refresh token is in Redis
	userID, err := repository.GetCache("refresh_token:" + req.RefreshToken)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to check refresh token", http.StatusInternalServerError, err))
		return
	}
	if userID == "" {
		c.Error(appErrors.NewAppError("Invalid refresh token", http.StatusUnauthorized, nil))
		return
	}

	// Invalidate old refresh token
	if err := repository.DeleteCache("refresh_token:" + req.RefreshToken); err != nil {
		c.Error(appErrors.NewAppError("Failed to invalidate old refresh token", http.StatusInternalServerError, err))
		return
	}

	var user domain.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
		return
	}

	accessToken, refreshToken, err := auth.GenerateTokens(user.ID, user.Role)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to generate tokens", http.StatusInternalServerError, err))
		return
	}

	accessTokenExpiresAt := time.Now().Add(8 * time.Hour)
	if err := repository.SetCache("access_token:"+accessToken, user.ID, accessTokenExpiresAt.Sub(time.Now())); err != nil {
		c.Error(appErrors.NewAppError("Failed to store access token", http.StatusInternalServerError, err))
		return
	}

	refreshTokenExpiresAt := time.Now().Add(10 * 365 * 24 * time.Hour)
	if err := repository.SetCache("refresh_token:"+refreshToken, user.ID, refreshTokenExpiresAt.Sub(time.Now())); err != nil {
		c.Error(appErrors.NewAppError("Failed to store refresh token", http.StatusInternalServerError, err))
		return
	}

	csrfToken, err := auth.GenerateCSRFToken(user.ID)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to generate CSRF token", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"csrfToken":    csrfToken,
	})
}

// LogoutUser godoc
// @Summary Log out a user
// @Description Invalidate both the access and refresh tokens
// @Tags users
// @Accept json
// @Produce json
// @Param token body requests.RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/logout [post]
func (h *UserHandler) LogoutUser(c *gin.Context) {
	var req requests.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	// Invalidate refresh token
	if err := repository.DeleteCache("refresh_token:" + req.RefreshToken); err != nil {
		c.Error(appErrors.NewAppError("Failed to invalidate refresh token", http.StatusInternalServerError, err))
		return
	}

	// Invalidate access token
	authHeader := c.GetHeader("Authorization")
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if err := repository.DeleteCache("access_token:" + tokenString); err != nil {
		// Even if the access token is already expired, we should proceed with logout
	}

	csrfToken := c.GetHeader("X-CSRF-Token")
	if err := auth.InvalidateCSRFToken(csrfToken); err != nil {
		c.Error(appErrors.NewAppError("Failed to invalidate CSRF token", http.StatusInternalServerError, err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// GetUser godoc
// @Summary Get user details by ID
// @Description Get details of a specific user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Security ApiKeyAuth
// @Success 200 {object} domain.User
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	var user domain.User

	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("User not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
		return
	}

	if !canAccessUser(c, user.ID) {
		c.Error(appErrors.NewAppError("Forbidden", http.StatusForbidden, nil))
		return
	}
	// Do not return hashed password
	user.Password = ""
	c.JSON(http.StatusOK, user)
}

// UpdateUser godoc
// @Summary Update user details
// @Description Update details of a specific user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body requests.UserUpdateRequest true "User update request"
// @Security ApiKeyAuth
// @Success 200 {object} domain.User
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req requests.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	var user domain.User
	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("User not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
		return
	}

	// Get the role of the authenticated user from the context
	authUserRole, exists := c.Get("role")
	if !exists {
		c.Error(appErrors.NewAppError("User role not found in context", http.StatusInternalServerError, nil))
		return
	}

	// If the request includes a role change, ensure the authenticated user is an Admin
	if req.Role != "" && authUserRole != "Admin" {
		c.Error(appErrors.NewAppError("Only Admins can change user roles", http.StatusForbidden, nil))
		return
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.Error(appErrors.NewAppError("Failed to hash password", http.StatusInternalServerError, err))
			return
		}
		user.Password = string(hashedPassword)
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	if err := h.userRepo.UpdateUser(&user); err != nil {
		c.Error(appErrors.NewAppError("Failed to update user", http.StatusInternalServerError, err))
		return
	}

	user.Password = "" // Don't return password
	c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a specific user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Security ApiKeyAuth
// @Success 204 "No Content"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user domain.User

	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("User not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
		return
	}

	if err := h.userRepo.DeleteUser(&user); err != nil {
		c.Error(appErrors.NewAppError("Failed to delete user", http.StatusInternalServerError, err))
		return
	}

	c.Status(http.StatusNoContent)
}

// ApproveUser godoc
// @Summary Approve a user
// @Description Activate a user's account by setting IsActive to true
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Security ApiKeyAuth
// @Success 200 {object} domain.User
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id}/approve [put]
func (h *UserHandler) ApproveUser(c *gin.Context) {
	id := c.Param("id")
	var user domain.User

	if err := h.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("User not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
		return
	}

	user.IsActive = true
	if err := h.userRepo.UpdateUser(&user); err != nil {
		c.Error(appErrors.NewAppError("Failed to approve user", http.StatusInternalServerError, err))
		return
	}

	user.Password = ""
	c.JSON(http.StatusOK, user)
}

func canAccessUser(c *gin.Context, targetUserID uint) bool {
	roleVal, roleExists := c.Get("role")
	requestorVal, userExists := c.Get("user_id")
	if !roleExists || !userExists {
		return false
	}

	role, _ := roleVal.(string)
	requestorID, _ := requestorVal.(uint)

	if role == "Admin" {
		return true
	}

	return requestorID == targetUserID
}
