package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"inventory/backend/internal/auth"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
)

// AuthMiddleware authenticates requests using JWT.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Error(appErrors.NewAppError("Authorization header required", http.StatusUnauthorized, nil))
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader { // No "Bearer " prefix found
			c.Error(appErrors.NewAppError("Bearer token required", http.StatusUnauthorized, nil))
			c.Abort()
			return
		}

		// Check if token is in Redis
		userID, err := repository.GetCache("access_token:" + tokenString)
		if err != nil {
			c.Error(appErrors.NewAppError("Failed to check access token", http.StatusInternalServerError, err))
			c.Abort()
			return
		}
		if userID == "" {
			c.Error(appErrors.NewAppError("Invalid access token", http.StatusUnauthorized, nil))
			c.Abort()
			return
		}

		claims, err := auth.ValidateJWT(tokenString)
		if err != nil {
			c.Error(appErrors.NewAppError("Invalid token", http.StatusUnauthorized, err))
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// AuthorizeMiddleware checks if the user has one of the required roles.
func AuthorizeMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.Error(appErrors.NewAppError("Role information not found", http.StatusForbidden, nil))
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			c.Error(appErrors.NewAppError("Invalid role format", http.StatusInternalServerError, nil))
			c.Abort()
			return
		}

		for _, requiredRole := range requiredRoles {
			if roleStr == requiredRole {
				c.Next()
				return
			}
		}

		c.Error(appErrors.NewAppError("Insufficient permissions", http.StatusForbidden, nil))
		c.Abort()
	}
}
