package repository

import (
	"context"
	"fmt"
	"inventory/backend/internal/domain"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type RoleRepository interface {
	ListRoles() ([]domain.Role, error)
	GetRoleByID(id uint) (*domain.Role, error)
	GetRoleByName(name string) (*domain.Role, error)
	CreateRole(role *domain.Role) error
	UpdateRole(role *domain.Role) error
	DeleteRole(id uint) error
	ListPermissions() ([]domain.Permission, error)
	GetPermissionsByRoleID(roleID uint) ([]domain.Permission, error)
	GetPermissionsByRoleName(roleName string) ([]domain.Permission, error)
	AssignPermissions(roleID uint, permissionIDs []uint) error
}

type roleRepository struct {
	db    *gorm.DB
	redis redis.UniversalClient
}

func NewRoleRepository(db *gorm.DB, redisClient redis.UniversalClient) RoleRepository {
	return &roleRepository{db: db, redis: redisClient}
}

func (r *roleRepository) ListRoles() ([]domain.Role, error) {
	var roles []domain.Role
	// Preload permissions count or details if needed. For list, maybe just count.
	// But specific requirement might need list of perms.
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

func (r *roleRepository) GetRoleByID(id uint) (*domain.Role, error) {
	var role domain.Role
	err := r.db.Preload("Permissions").First(&role, id).Error
	return &role, err
}

func (r *roleRepository) GetRoleByName(name string) (*domain.Role, error) {
	var role domain.Role
	err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error
	return &role, err
}

func (r *roleRepository) CreateRole(role *domain.Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) UpdateRole(role *domain.Role) error {
	// Invalidate cache before/after update
	// ideally we know the OLD name if it changed, but for now just invalidate current name in struct
	// If name changes, the middleware won't find the old key anyway.
	if err := r.db.Save(role).Error; err != nil {
		return err
	}
	// Invalidate cache
	if r.redis != nil {
		ctx := context.Background()
		r.redis.Del(ctx, fmt.Sprintf("rbac:role:%s:permissions", role.Name))
	}
	return nil
}

func (r *roleRepository) DeleteRole(id uint) error {
	// Get role first to invalidate cache
	var role domain.Role
	if err := r.db.First(&role, id).Error; err != nil {
		return err // Or ignore if not found
	}

	if err := r.db.Delete(&domain.Role{}, id).Error; err != nil {
		return err
	}

	if r.redis != nil {
		ctx := context.Background()
		r.redis.Del(ctx, fmt.Sprintf("rbac:role:%s:permissions", role.Name))
	}
	return nil
}

func (r *roleRepository) ListPermissions() ([]domain.Permission, error) {
	var perms []domain.Permission
	// Order by group for easier frontend display
	err := r.db.Order("\"group\", name").Find(&perms).Error
	return perms, err
}

func (r *roleRepository) GetPermissionsByRoleID(roleID uint) ([]domain.Permission, error) {
	var role domain.Role
	if err := r.db.Preload("Permissions").First(&role, roleID).Error; err != nil {
		return nil, err
	}
	return role.Permissions, nil
}

func (r *roleRepository) AssignPermissions(roleID uint, permissionIDs []uint) error {
	var role domain.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}

	var perms []domain.Permission
	if len(permissionIDs) > 0 {
		if err := r.db.Where("id IN ?", permissionIDs).Find(&perms).Error; err != nil {
			return err
		}
	}

	// Helper to replace association
	if err := r.db.Model(&role).Association("Permissions").Replace(perms); err != nil {
		return err
	}

	// Invalidate Cache
	if r.redis != nil {
		ctx := context.Background()
		r.redis.Del(ctx, fmt.Sprintf("rbac:role:%s:permissions", role.Name))
	}
	return nil
}
