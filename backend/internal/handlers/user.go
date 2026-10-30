package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"inventory/backend/internal/auth"
	"inventory/backend/internal/config"
	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
)

type UserHandler struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

func NewUserHandler(userRepo repository.UserRepository, cfg *config.Config) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func resolveRefreshTTL(cfg *config.Config) time.Duration {
	if cfg == nil || cfg.RefreshTokenTTL <= 0 {
		return 720 * time.Hour
	}
	return cfg.RefreshTokenTTL
}

func (h *UserHandler) secureCookies() bool {
	if h.cfg == nil {
		return false
	}
	return h.secureCookies()
}

// ListUsers godoc
// @Summary List users with optional status and search filters
// @Description Retrieves all users, optionally filtered by status (approved/pending) and search query (username or ID). Supports pagination.
// @Tags users
// @Accept json
// @Produce json
// @Param status query string false "Filter by user status (approved, pending)"
// @Param q query string false "Search by username or ID"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 50, max: 200)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Paginated list of users"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit > 200 {
		limit = 200 // Cap at 200 to prevent abuse
	}
	status := strings.ToLower(c.Query("status"))
	search := c.Query("q")

	users, total, err := h.userRepo.ListUsers(page, limit, status, search)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to list users", http.StatusInternalServerError, err))
		return
	}

	// Remove passwords from response
	for i := range users {
		users[i].Password = ""
	}

	// Return paginated response
	response := gin.H{
		"users":        users,
		"totalItems":   total,
		"currentPage":  page,
		"totalPages":   (total + int64(limit) - 1) / int64(limit),
		"itemsPerPage": limit,
	}

	c.JSON(http.StatusOK, response)
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
	userCount, err := h.userRepo.CountUsers()
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to check for existing users", http.StatusInternalServerError, err))
		return
	}

	var user domain.User
	var role *domain.Role

	// Lookup the role
	roleName := req.Role
	// Enforce Admin for first user
	if userCount <= 1 {
		roleName = "Admin"
		if req.Role != "Admin" {
			c.Error(appErrors.NewAppError("The first user must be an Admin", http.StatusBadRequest, nil))
			return
		}
	}

	role, err = h.userRepo.GetRoleByName(roleName)
	if err != nil {
		c.Error(appErrors.NewAppError("Invalid role", http.StatusBadRequest, err))
		return
	}

	user = domain.User{
		Username:    req.Username,
		RoleID:      role.ID,
		Role:        *role,
		RoleName:    role.Name,
		LegacyRole:  role.Name, // Populate legacy column
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Address:     req.Address,
		IsActive:    userCount <= 1, // First human user is active by default (AI agent is 0)
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

	var user *domain.User
	// Preload Role AND Permissions
	user, err := h.userRepo.GetUserByUsername(req.Username)
	if err != nil {
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

	// Dynamic RBAC Migration: LegacyRole removed.
	// if user.RoleID == 0 && user.LegacyRole != "" { ... }

	// Fallback/Safety: If still no role, try to default to Customer or fail
	if user.Role.ID == 0 {
		c.Error(appErrors.NewAppError("User has no role assigned. Please contact support.", http.StatusForbidden, nil))
		return
	}

	accessToken, refreshToken, accessJTI, err := auth.GenerateTokens(user.ID, user.Role.Name)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to generate tokens", http.StatusInternalServerError, err))
		return
	}

	// Store tokens in Redis using JTI for Access Token
	accessTokenExpiresAt := time.Now().Add(8 * time.Hour)
	if err := repository.SetCache("access_token:"+accessJTI, user.ID, time.Until(accessTokenExpiresAt)); err != nil {
		c.Error(appErrors.NewAppError("Failed to store access token", http.StatusInternalServerError, err))
		return
	}

	refreshTokenExpiresAt := time.Now().Add(resolveRefreshTTL(h.cfg))
	if err := repository.SetCache("refresh_token:"+refreshToken, user.ID, time.Until(refreshTokenExpiresAt)); err != nil {
		c.Error(appErrors.NewAppError("Failed to store refresh token", http.StatusInternalServerError, err))
		return
	}

	csrfToken, err := auth.GenerateCSRFToken(user.ID)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to generate CSRF token", http.StatusInternalServerError, err))
		return
	}

	// Extract permissions
	var permissions []string
	if user.Role.Permissions != nil {
		for _, p := range user.Role.Permissions {
			permissions = append(permissions, p.Name)
		}
	}

	logrus.WithFields(logrus.Fields{
		"role":        user.Role.Name,
		"role_id":     user.Role.ID,
		"permissions": len(permissions),
	}).Debug("LoginUser resolved role permissions")

	auth.SetSessionCookies(c, accessToken, refreshToken, csrfToken, resolveRefreshTTL(h.cfg), h.secureCookies())

	user.Password = ""
	// Tokens live only in httpOnly cookies; the SPA receives identity and the
	// non-secret CSRF token needed for the double-submit header.
	c.JSON(http.StatusOK, gin.H{
		"csrfToken":   csrfToken,
		"user":        user,
		"permissions": permissions,
		"expiresIn":   int(auth.AccessTokenTTL.Seconds()),
	})
}

// GetCurrentUser godoc
// @Summary Get the authenticated user
// @Description Returns the current user and effective permissions. Used by the
// SPA to validate its cookie session after a reload.
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /users/me [get]
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.Error(appErrors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
		return
	}
	uid, ok := userID.(uint)
	if !ok {
		c.Error(appErrors.NewAppError("Invalid user context", http.StatusInternalServerError, nil))
		return
	}

	user, err := h.userRepo.GetUserByID(strconv.FormatUint(uint64(uid), 10))
	if err != nil {
		c.Error(appErrors.NewAppError("User not found", http.StatusUnauthorized, err))
		return
	}

	permissions := make([]string, 0, len(user.Role.Permissions))
	for _, p := range user.Role.Permissions {
		permissions = append(permissions, p.Name)
	}

	// Ensure the SPA always has a usable CSRF token, even when the API and SPA
	// are on different subdomains and the cookie is not JS-readable.
	csrfToken := auth.CSRFTokenFromRequest(c)
	if csrfToken == "" {
		csrfToken, _ = auth.GenerateCSRFToken(uid)
	} else if valid, _ := auth.ValidateCSRFToken(csrfToken, uid); !valid {
		csrfToken, _ = auth.GenerateCSRFToken(uid)
	}
	if csrfToken != "" {
		auth.SetCSRFCookie(c, csrfToken, h.secureCookies())
	}

	user.Password = ""
	c.JSON(http.StatusOK, gin.H{"user": user, "permissions": permissions, "csrfToken": csrfToken})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Get a new access token using a refresh token
// @Tags users
// @Accept json
// @Produce json
// @Param token body requests.RefreshTokenRequest true "Refresh token request"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "New access token"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized: Invalid refresh token"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/refresh-token [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req requests.RefreshTokenRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(appErrors.NewAppError("Invalid request payload", http.StatusBadRequest, err))
			return
		}
	}

	refreshToken := auth.RefreshTokenFromRequest(c, req.RefreshToken)
	if refreshToken == "" {
		c.Error(appErrors.NewAppError("Refresh token required", http.StatusUnauthorized, nil))
		return
	}

	// Check if refresh token is in Redis
	userID, err := repository.GetCache("refresh_token:" + refreshToken)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to check refresh token", http.StatusInternalServerError, err))
		return
	}
	if userID == "" {
		c.Error(appErrors.NewAppError("Invalid refresh token", http.StatusUnauthorized, nil))
		return
	}

	// Invalidate old refresh token
	if err := repository.DeleteCache("refresh_token:" + refreshToken); err != nil {
		c.Error(appErrors.NewAppError("Failed to invalidate old refresh token", http.StatusInternalServerError, err))
		return
	}

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
		return
	}

	accessToken, refreshToken, accessJTI, err := auth.GenerateTokens(user.ID, user.Role.Name)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to generate tokens", http.StatusInternalServerError, err))
		return
	}

	accessTokenExpiresAt := time.Now().Add(8 * time.Hour)
	if err := repository.SetCache("access_token:"+accessJTI, user.ID, time.Until(accessTokenExpiresAt)); err != nil {
		c.Error(appErrors.NewAppError("Failed to store access token", http.StatusInternalServerError, err))
		return
	}

	refreshTokenExpiresAt := time.Now().Add(resolveRefreshTTL(h.cfg))
	if err := repository.SetCache("refresh_token:"+refreshToken, user.ID, time.Until(refreshTokenExpiresAt)); err != nil {
		c.Error(appErrors.NewAppError("Failed to store refresh token", http.StatusInternalServerError, err))
		return
	}

	if oldCSRF := auth.CSRFTokenFromRequest(c); oldCSRF != "" {
		_ = auth.InvalidateCSRFToken(oldCSRF)
	}
	csrfToken, err := auth.GenerateCSRFToken(user.ID)
	if err != nil {
		c.Error(appErrors.NewAppError("Failed to generate CSRF token", http.StatusInternalServerError, err))
		return
	}

	auth.SetSessionCookies(c, accessToken, refreshToken, csrfToken, resolveRefreshTTL(h.cfg), h.secureCookies())
	c.JSON(http.StatusOK, gin.H{
		"csrfToken": csrfToken,
		"expiresIn": int(auth.AccessTokenTTL.Seconds()),
	})
}

// LogoutUser godoc
// @Summary Log out a user
// @Description Invalidate both the access and refresh tokens
// @Tags users
// @Accept json
// @Produce json
// @Param token body requests.RefreshTokenRequest true "Refresh token request"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/logout [post]
func (h *UserHandler) LogoutUser(c *gin.Context) {
	var req requests.RefreshTokenRequest
	if c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&req) // body is optional; cookies carry the session
	}

	// Invalidate refresh token (cookie or explicit body token).
	if refreshToken := auth.RefreshTokenFromRequest(c, req.RefreshToken); refreshToken != "" {
		if err := repository.DeleteCache("refresh_token:" + refreshToken); err != nil {
			c.Error(appErrors.NewAppError("Failed to invalidate refresh token", http.StatusInternalServerError, err))
			return
		}
	}

	// Invalidate access token (header or cookie).
	if tokenString := auth.AccessTokenFromRequest(c); tokenString != "" {
		if claims, err := auth.GetTokenClaims(tokenString); err == nil && claims.ID != "" {
			_ = repository.DeleteCache("access_token:" + claims.ID)
		}
	}

	if csrfToken := auth.CSRFTokenFromRequest(c); csrfToken != "" {
		_ = auth.InvalidateCSRFToken(csrfToken)
	}

	auth.ClearSessionCookies(c, h.secureCookies())
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

// GetUser godoc
// @Summary Get user details by ID
// @Description Get details of a specific user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Security BearerAuth
// @Success 200 {object} domain.User
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userRepo.GetUserByID(id)
	if err != nil {
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
// @Security BearerAuth
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

	user, err := h.userRepo.GetUserByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("User not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
		return
	}

	// Get the role of the authenticated user from the context
	authUserRoleVal, exists := c.Get("role")
	if !exists {
		c.Error(appErrors.NewAppError("User role not found in context", http.StatusInternalServerError, nil))
		return
	}

	authUserRole, _ := authUserRoleVal.(string)

	if !canAccessUser(c, user.ID) && authUserRole != "Admin" {
		c.Error(appErrors.NewAppError("Forbidden: cannot update this user", http.StatusForbidden, nil))
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
		role, err := h.userRepo.GetRoleByName(req.Role)
		if err != nil {
			c.Error(appErrors.NewAppError("Invalid role", http.StatusBadRequest, err))
			return
		}
		user.RoleID = role.ID
		user.Role = *role
		user.RoleName = role.Name
	}
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.PhoneNumber != "" {
		user.PhoneNumber = req.PhoneNumber
	}
	if req.Address != "" {
		user.Address = req.Address
	}

	if err := h.userRepo.UpdateUser(user); err != nil {
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
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userRepo.GetUserByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("User not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
		return
	}

	if err := h.userRepo.DeleteUser(user); err != nil {
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
// @Security BearerAuth
// @Success 200 {object} domain.User
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /users/{id}/approve [put]
func (h *UserHandler) ApproveUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.userRepo.GetUserByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Error(appErrors.NewAppError("User not found", http.StatusNotFound, err))
			return
		}
		c.Error(appErrors.NewAppError("Failed to fetch user", http.StatusInternalServerError, err))
		return
	}

	user.IsActive = true
	if err := h.userRepo.UpdateUser(user); err != nil {
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
