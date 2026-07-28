//go:build integration

package integration

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/repository"
)

func TestPermissionCacheInvalidation(t *testing.T) {
	db := SetupPostgres(t)
	Migrate(t, db)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	role := domain.Role{Name: "Staff", Description: "Standard access"}
	require.NoError(t, db.Create(&role).Error)

	perm1 := domain.Permission{Name: "products.read", Group: "Inventory"}
	perm2 := domain.Permission{Name: "products.write", Group: "Inventory"}
	require.NoError(t, db.Create(&perm1).Error)
	require.NoError(t, db.Create(&perm2).Error)

	roleRepo := repository.NewRoleRepository(db, redisClient)

	require.NoError(t, roleRepo.AssignPermissions(role.ID, []uint{perm1.ID}, 1))

	perms, err := roleRepo.GetPermissionsByRoleName("Staff")
	require.NoError(t, err)
	assert.Len(t, perms, 1)
	assert.Equal(t, "products.read", perms[0].Name)

	cacheKey := "rbac:role:Staff:permissions"
	assert.True(t, mr.Exists(cacheKey))

	require.NoError(t, roleRepo.AssignPermissions(role.ID, []uint{perm1.ID, perm2.ID}, 1))

	assert.False(t, mr.Exists(cacheKey))

	perms, err = roleRepo.GetPermissionsByRoleName("Staff")
	require.NoError(t, err)
	assert.Len(t, perms, 2)

	assert.True(t, mr.Exists(cacheKey))
}
