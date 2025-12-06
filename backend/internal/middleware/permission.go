package middleware

import (
	"inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequirePermission checks if the user's role has the specific permission.
// Admins are always allowed.
func RequirePermission(roleRepo repository.RoleRepository, requiredPerm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get Role Name from Context (set by AuthMiddleware)
		roleVal, exists := c.Get("role")
		if !exists {
			c.Error(errors.NewAppError("Role information not found", http.StatusForbidden, nil))
			c.Abort()
			return
		}
		roleName, ok := roleVal.(string)
		if !ok {
			c.Error(errors.NewAppError("Invalid role format", http.StatusInternalServerError, nil))
			c.Abort()
			return
		}

		// 2. Admin Bypass
		if roleName == "Admin" {
			c.Next()
			return
		}

		// 3. Fetch Permissions for Role
		// Optimization: This hits DB on every request. Consider caching role permissions in Redis.
		// For now, relies on Postgres performance.
		perms, err := roleRepo.GetPermissionsByRoleName(roleName)
		if err != nil {
			c.Error(errors.NewAppError("Failed to verify permissions", http.StatusInternalServerError, err))
			c.Abort()
			return
		}

		// 4. Check if requiredPerm exists
		for _, p := range perms {
			if p.Name == requiredPerm {
				c.Next()
				return
			}
		}

		c.Error(errors.NewAppError("Insufficient permissions: required "+requiredPerm, http.StatusForbidden, nil))
		c.Abort()
	}
}
