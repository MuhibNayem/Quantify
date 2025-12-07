package handlers

import (
	"inventory/backend/internal/errors"
	"inventory/backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	service services.RoleService
}

func NewRoleHandler(service services.RoleService) *RoleHandler {
	return &RoleHandler{service: service}
}

// ListRoles godoc
// @Summary List all roles
// @Tags roles
// @Produce json
// @Success 200 {array} domain.Role
// @Router /roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	roles, err := h.service.ListRoles()
	if err != nil {
		c.Error(errors.NewAppError("Failed to list roles", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, roles)
}

// CreateRole godoc
// @Summary Create a new role
// @Tags roles
// @Accept json
// @Produce json
// @Param role body object true "Role data"
// @Success 201 {object} domain.Role
// @Router /roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewAppError("Invalid request", http.StatusBadRequest, err))
		return
	}

	role, err := h.service.CreateRole(req.Name, req.Description)
	if err != nil {
		c.Error(errors.NewAppError("Failed to create role", http.StatusBadRequest, err))
		return
	}
	c.JSON(http.StatusCreated, role)
}

// UpdateRole godoc
// @Summary Update role details
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param role body object true "Role data"
// @Success 200 {object} domain.Role
// @Router /roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewAppError("Invalid request", http.StatusBadRequest, err))
		return
	}

	role, err := h.service.UpdateRole(uint(id), req.Name, req.Description)
	if err != nil {
		c.Error(errors.NewAppError("Failed to update role", http.StatusBadRequest, err))
		return
	}
	c.JSON(http.StatusOK, role)
}

// DeleteRole godoc
// @Summary Delete a custom role
// @Tags roles
// @Param id path int true "Role ID"
// @Success 204 "No content"
// @Router /roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.DeleteRole(uint(id)); err != nil {
		c.Error(errors.NewAppError("Failed to delete role", http.StatusBadRequest, err))
		return
	}
	c.Status(http.StatusNoContent)
}

// ListPermissions godoc
// @Summary List all available system permissions
// @Tags roles
// @Produce json
// @Success 200 {array} domain.Permission
// @Router /permissions [get]
func (h *RoleHandler) ListPermissions(c *gin.Context) {
	perms, err := h.service.ListPermissions()
	if err != nil {
		c.Error(errors.NewAppError("Failed to list permissions", http.StatusInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, perms)
}

// UpdateRolePermissions godoc
// @Summary Assign permissions to a role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param body body object true "Permission IDs"
// @Success 200 "OK"
// @Router /roles/{id}/permissions [put]
func (h *RoleHandler) UpdateRolePermissions(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		PermissionIDs []uint `json:"permission_ids"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewAppError("Invalid request", http.StatusBadRequest, err))
		return
	}

	if err := h.service.UpdateRolePermissions(uint(id), req.PermissionIDs); err != nil {
		c.Error(errors.NewAppError("Failed to update permissions", http.StatusBadRequest, err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
