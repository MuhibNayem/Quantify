package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appErrors "inventory/backend/internal/errors" // Alias to avoid conflict with standard errors
)

// ErrorHandler is a middleware to handle custom application errors.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Process request

		// Check if an error occurred during request processing
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				// Check if it's our custom AppError
				if appErr, ok := e.Err.(*appErrors.AppError); ok {
					c.JSON(appErr.StatusCode, gin.H{"error": appErr.Message})
					return
				}
				// For any other unhandled errors, return a generic internal server error
				c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred"})
				return
			}
		}
	}
}
