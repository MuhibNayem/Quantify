package middleware

import (
	"inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RequirePermission checks if the user's role has the specific permission.
// Admins are always allowed.
func RequirePermission(roleRepo repository.RoleRepository, requiredPerm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !permissionAllowed(c, roleRepo, requiredPerm) {
			c.Abort()
			return
		}
		c.Next()
	}
}

// SelfOrPermission allows the request when the :userId route parameter matches
// the authenticated user, otherwise falls back to a permission check. It is
// used for self-service resources (time tracking) that managers can also view.
func SelfOrPermission(roleRepo repository.RoleRepository, requiredPerm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.Error(errors.NewAppError("Authenticated user not found", http.StatusUnauthorized, nil))
			c.Abort()
			return
		}

		if param := c.Param("userId"); param != "" {
			if uid, ok := userID.(uint); ok && strconv.FormatUint(uint64(uid), 10) == param {
				c.Next()
				return
			}
		}

		if !permissionAllowed(c, roleRepo, requiredPerm) {
			c.Abort()
			return
		}
		c.Next()
	}
}

// permissionAllowed evaluates the role/permission pair and records an AppError
// when access is denied.
func permissionAllowed(c *gin.Context, roleRepo repository.RoleRepository, requiredPerm string) bool {
	roleVal, exists := c.Get("role")
	if !exists {
		c.Error(errors.NewAppError("Role information not found", http.StatusForbidden, nil))
		return false
	}
	roleName, ok := roleVal.(string)
	if !ok {
		c.Error(errors.NewAppError("Invalid role format", http.StatusInternalServerError, nil))
		return false
	}

	// Admin bypass.
	if roleName == "Admin" {
		return true
	}

	// Permissions are cached in Redis with a 60s TTL and invalidated on every
	// role mutation by the repository helpers.
	perms, err := roleRepo.GetPermissionsByRoleName(roleName)
	if err != nil {
		c.Error(errors.NewAppError("Failed to verify permissions", http.StatusInternalServerError, err))
		return false
	}

	for _, p := range perms {
		if p.Name == requiredPerm {
			return true
		}
	}

	c.Error(errors.NewAppError("Insufficient permissions: required "+requiredPerm, http.StatusForbidden, nil))
	return false
}
