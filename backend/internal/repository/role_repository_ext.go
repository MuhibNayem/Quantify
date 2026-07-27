package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"inventory/backend/internal/domain"
)

const (
	permissionCacheKeyFmt = "rbac:role:%s:permissions"
	PermissionCacheTTL    = 60 * time.Second
)

// csrfTokenKeyPrefix must stay in sync with auth/csrf.go (key format is "csrf_token:<token>").
const csrfTokenKeyPrefix = "csrf_token:"
const csrfScanBatch = 200

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

// rotateCSRFForRole deletes every CSRF token bound to any user that currently
// holds the supplied role. It mirrors auth.RotateCSRFTokenFor but is
// duplicated here to avoid an import cycle with the auth package.
//
// Failures are logged and swallowed: a CSRF rotation miss must never block a
// legitimate privilege change, but it is a security signal worth surfacing.
func (r *roleRepository) rotateCSRFForRole(ctx context.Context, roleID uint) {
	if r.redis == nil {
		return
	}

	var userIDs []uint
	if err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("role_id = ?", roleID).
		Pluck("id", &userIDs).Error; err != nil {
		logrus.WithError(err).WithField("role_id", roleID).
			Warn("CSRF rotation: failed to list users with role")
		return
	}
	if len(userIDs) == 0 {
		return
	}

	targets := make(map[string]struct{}, len(userIDs))
	for _, uid := range userIDs {
		targets[strconv.FormatUint(uint64(uid), 10)] = struct{}{}
	}

	iter := r.redis.Scan(ctx, 0, csrfTokenKeyPrefix+"*", csrfScanBatch).Iterator()
	keys := make([]string, 0, csrfScanBatch)
	for iter.Next(ctx) {
		key := iter.Val()
		owner, err := r.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		if _, match := targets[owner]; match {
			keys = append(keys, key)
		}
	}
	if err := iter.Err(); err != nil {
		logrus.WithError(err).WithField("role_id", roleID).
			Warn("CSRF rotation: scan failed")
		return
	}
	if len(keys) == 0 {
		return
	}
	if err := r.redis.Del(ctx, keys...).Err(); err != nil {
		logrus.WithError(err).WithField("role_id", roleID).
			Warn("CSRF rotation: delete failed")
	}
}

// writeRolePermissionsAudit persists an audit_log row recording the
// old→new permission set for a role grant. Failures are swallowed and
// logged by WriteAudit; the grant itself is not blocked by audit issues.
func (r *roleRepository) writeRolePermissionsAudit(ctx context.Context, roleID, userID uint, oldNames, newNames []string) {
	changes, err := json.Marshal(map[string][]string{
		"old": oldNames,
		"new": newNames,
	})
	if err != nil {
		logrus.WithError(err).WithField("role_id", roleID).
			Warn("role permission audit: failed to marshal changes")
		return
	}
	entry := &domain.AuditLog{
		Action:    "UPDATE",
		Entity:    "Role",
		EntityID:  strconv.FormatUint(uint64(roleID), 10),
		UserID:    userID,
		Changes:   string(changes),
		Timestamp: time.Now(),
	}
	_ = WriteAudit(ctx, r.db, entry)
}
