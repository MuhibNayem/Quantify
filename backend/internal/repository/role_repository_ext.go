package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"inventory/backend/internal/domain"
)

const (
	permissionCacheKeyFmt = "rbac:role:%s:permissions"
	PermissionCacheTTL    = 60 * time.Second
)

func permissionCacheKey(roleName string) string {
	return fmt.Sprintf(permissionCacheKeyFmt, roleName)
}

func (r *roleRepository) SetPermissionsCache(ctx context.Context, roleName string, perms []string) error {
	if r.redis == nil {
		return nil
	}
	data, err := json.Marshal(perms)
	if err != nil {
		return fmt.Errorf("marshal permissions for role %q: %w", roleName, err)
	}
	if err := r.redis.Set(ctx, permissionCacheKey(roleName), data, PermissionCacheTTL).Err(); err != nil {
		logrus.WithError(err).WithField("role", roleName).Warn("failed to set permission cache")
		return err
	}
	return nil
}

func (r *roleRepository) InvalidateRoleCache(ctx context.Context, roleName string) error {
	if r.redis == nil {
		return nil
	}
	if err := r.redis.Del(ctx, permissionCacheKey(roleName)).Err(); err != nil {
		logrus.WithError(err).WithField("role", roleName).Warn("failed to invalidate permission cache")
		return err
	}
	return nil
}

func (r *roleRepository) GetPermissionsByRoleName(roleName string) ([]domain.Permission, error) {
	ctx := context.Background()
	cacheKey := permissionCacheKey(roleName)

	if r.redis != nil {
		cached, err := r.redis.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var names []string
			if err := json.Unmarshal(cached, &names); err == nil {
				return namesToPermissions(names), nil
			}
			logrus.WithError(err).WithField("role", roleName).Warn("permission cache corrupt, refetching")
		} else if err.Error() != "redis: nil" {
			logrus.WithError(err).WithField("role", roleName).Warn("permission cache read failed")
		}
	}

	names, err := r.loadPermissionNamesByRoleName(ctx, roleName)
	if err != nil {
		return nil, err
	}
	if r.redis != nil {
		if err := r.SetPermissionsCache(ctx, roleName, names); err != nil {
			logrus.WithError(err).WithField("role", roleName).Warn("permission cache write failed")
		}
	}
	return namesToPermissions(names), nil
}

func (r *roleRepository) loadPermissionNamesByRoleName(ctx context.Context, roleName string) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).
		Table("permissions p").
		Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
		Joins("JOIN roles r ON r.id = rp.role_id").
		Where("r.name = ?", roleName).
		Pluck("p.name", &names).Error
	if err != nil {
		return nil, err
	}
	return names, nil
}

func namesToPermissions(names []string) []domain.Permission {
	perms := make([]domain.Permission, 0, len(names))
	for _, n := range names {
		perms = append(perms, domain.Permission{Name: n})
	}
	return perms
}
