package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	appErrors "inventory/backend/internal/errors"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		for _, e := range c.Errors {
			fields := logrus.Fields{
				"path":   c.FullPath(),
				"method": c.Request.Method,
			}
			if fields["path"] == "" {
				fields["path"] = c.Request.URL.Path
			}
			if userID, exists := c.Get("user_id"); exists {
				fields["user_id"] = userID
			}

			if appErr, ok := e.Err.(*appErrors.AppError); ok {
				logrus.WithFields(fields).WithError(appErr.Err).Error(appErr.Message)
				c.JSON(appErr.StatusCode, gin.H{"error": appErr.Message})
				return
			}

			logrus.WithFields(fields).WithError(e.Err).Error("Unhandled error")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error occurred"})
			return
		}
	}
}
