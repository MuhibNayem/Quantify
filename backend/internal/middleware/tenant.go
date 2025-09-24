package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
)

// TenantMiddleware extracts tenant ID from header and sets it in context.
func TenantMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantIDStr := c.GetHeader("X-Tenant-ID")
		if tenantIDStr == "" {
			c.Error(appErrors.NewAppError("X-Tenant-ID header is required", http.StatusBadRequest, nil))
			c.Abort()
			return
		}

		tenantID, err := strconv.ParseUint(tenantIDStr, 10, 64)
		if err != nil {
			c.Error(appErrors.NewAppError("Invalid X-Tenant-ID header", http.StatusBadRequest, err))
			c.Abort()
			return
		}

		var tenant domain.Tenant
		if err := repository.DB.First(&tenant, tenantID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.Error(appErrors.NewAppError("Tenant not found", http.StatusNotFound, err))
				c.Abort()
				return
			}
			c.Error(appErrors.NewAppError("Failed to fetch tenant", http.StatusInternalServerError, err))
			c.Abort()
			return
		}

		c.Set("tenant_id", uint(tenantID))
		c.Next()
	}
}

// GetTenantIDFromContext retrieves the tenant ID from the Gin context.
func GetTenantIDFromContext(c *gin.Context) (uint, error) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return 0, appErrors.NewAppError("Tenant ID not found in context", http.StatusInternalServerError, nil)
	}
	id, ok := tenantID.(uint)
	if !ok {
		return 0, appErrors.NewAppError("Invalid tenant ID format in context", http.StatusInternalServerError, nil)
	}
	return id, nil
}
