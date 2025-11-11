package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"inventory/backend/internal/auth"
	appErrors "inventory/backend/internal/errors"
)

var unsafeHTTPMethods = map[string]bool{
	http.MethodPost:   true,
	http.MethodPut:    true,
	http.MethodPatch:  true,
	http.MethodDelete: true,
}

// CSRFMiddleware enforces the presence of a valid CSRF token on unsafe HTTP methods.
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !unsafeHTTPMethods[c.Request.Method] {
			c.Next()
			return
		}

		token := c.GetHeader("X-CSRF-Token")
		if token == "" {
			c.Error(appErrors.NewAppError("Missing CSRF token", http.StatusForbidden, nil))
			c.Abort()
			return
		}

		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.Error(appErrors.NewAppError("Authenticated user not found in context", http.StatusUnauthorized, nil))
			c.Abort()
			return
		}

		userID, ok := userIDVal.(uint)
		if !ok {
			c.Error(appErrors.NewAppError("Invalid user ID type in context", http.StatusInternalServerError, nil))
			c.Abort()
			return
		}

		valid, err := auth.ValidateCSRFToken(token, userID)
		if err != nil {
			c.Error(appErrors.NewAppError("Failed to validate CSRF token", http.StatusInternalServerError, err))
			c.Abort()
			return
		}
		if !valid {
			c.Error(appErrors.NewAppError("Invalid CSRF token", http.StatusForbidden, nil))
			c.Abort()
			return
		}

		c.Next()
	}
}
