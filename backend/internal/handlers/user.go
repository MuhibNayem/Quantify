package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"inventory/backend/internal/auth"
	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/middleware"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
)

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
func RegisterUser(c *gin.Context) {
	var req requests.UserRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to hash password", http.StatusInternalServerError, err))
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(c)
	if err != nil {
		c.Error(err.(*appErrors.AppError))
		return
	}

	user := domain.User{
		TenantID: tenantID,
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     req.Role,
	}

	if err := repository.DB.Create(&user).Error; err != nil {
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
func LoginUser(c *gin.Context) {
	var req requests.UserLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
		return
	}

	tenantID, err := middleware.GetTenantIDFromContext(c)
	if err != nil {
		c.Error(err.(*appErrors.AppError))
		return
	}

	var user domain.User
	if err := repository.DB.Where("tenant_id = ? AND username = ?", tenantID, req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("Invalid credentials", http.StatusUnauthorized, nil))
			return
		}
		c.Error(appErrors.NewAppError("Database error", http.StatusInternalServerError, err))
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := auth.GenerateTokens(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Store tokens in Redis
	accessTokenExpiresAt := time.Now().Add(8 * time.Hour)
	if err := repository.SetCache("access_token:"+accessToken, user.ID, accessTokenExpiresAt.Sub(time.Now())); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store access token"})
		return
	}

	refreshTokenExpiresAt := time.Now().Add(10 * 365 * 24 * time.Hour) // 10 years
	if err := repository.SetCache("refresh_token:"+refreshToken, user.ID, refreshTokenExpiresAt.Sub(time.Now())); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
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
func RefreshToken(c *gin.Context) {
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
	if err := repository.DB.First(&user, userID).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
		return
	}

	// Generate new tokens
	accessToken, refreshToken, err := auth.GenerateTokens(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Store new tokens in Redis
	accessTokenExpiresAt := time.Now().Add(8 * time.Hour)
	if err := repository.SetCache("access_token:"+accessToken, user.ID, accessTokenExpiresAt.Sub(time.Now())); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store access token"})
		return
	}

	refreshTokenExpiresAt := time.Now().Add(10 * 365 * 24 * time.Hour) // 10 years
	if err := repository.SetCache("refresh_token:"+refreshToken, user.ID, refreshTokenExpiresAt.Sub(time.Now())); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
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
func LogoutUser(c *gin.Context) {
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
func GetUser(c *gin.Context) {
	id := c.Param("id")
	var user domain.User
	tenantID, err := middleware.GetTenantIDFromContext(c)
	if err != nil {
		c.Error(err.(*appErrors.AppError))
		return
	}

	if err := repository.DB.Where("tenant_id = ?", tenantID).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("User not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
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
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req requests.UserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user domain.User
	tenantID, err := middleware.GetTenantIDFromContext(c)
	if err != nil {
		c.Error(err.(*appErrors.AppError))
		return
	}

	if err := repository.DB.Where("tenant_id = ?", tenantID).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("User not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
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

	if err := repository.DB.Save(&user).Error; err != nil {
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
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user domain.User
	tenantID, err := middleware.GetTenantIDFromContext(c)
	if err != nil {
		c.Error(err.(*appErrors.AppError))
		return
	}

	if err := repository.DB.Where("tenant_id = ?", tenantID).First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("User not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
		return
	}

	if err := repository.DB.Delete(&user).Error; err != nil {
		c.Error(appErrors.NewAppError("Failed to delete user", http.StatusInternalServerError, err))
		return
	}

	c.Status(http.StatusNoContent)
}
