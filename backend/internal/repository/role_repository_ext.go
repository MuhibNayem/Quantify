package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"inventory/backend/internal/domain"
)

func (r *roleRepository) GetPermissionsByRoleName(roleName string) ([]domain.Permission, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("rbac:role:%s:permissions", roleName)

	// 1. Check Cache
	if r.redis != nil {
		cached, err := r.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var perms []domain.Permission
			if err := json.Unmarshal([]byte(cached), &perms); err == nil {
				return perms, nil
			}
		}
	}

	// 2. Database Lookup
	var role domain.Role
	// Preload Permissions
	if err := r.db.Preload("Permissions").Where("name = ?", roleName).First(&role).Error; err != nil {
		return nil, err
	}

	// 3. Set Cache
	if r.redis != nil {
		data, err := json.Marshal(role.Permissions)
		if err == nil {
			r.redis.Set(ctx, cacheKey, string(data), time.Hour)
		}
	}

	return role.Permissions, nil
}
