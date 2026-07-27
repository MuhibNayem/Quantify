package repository

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func setupRoleRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&domain.Role{},
		&domain.Permission{},
		&domain.RolePermission{},
		&domain.AuditLog{},
	))
	return db
}

func newRoleRepoWithRedis(t *testing.T, db *gorm.DB) (*roleRepository, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return &roleRepository{db: db, redis: rdb}, mr
}

func TestAssignPermissions_WritesAuditRow(t *testing.T) {
	db := setupRoleRepoTestDB(t)
	repo, _ := newRoleRepoWithRedis(t, db)

	role := domain.Role{Name: "Manager", IsSystem: false}
	require.NoError(t, db.Create(&role).Error)
	oldPerm := domain.Permission{Name: "products.read", Group: "Inventory"}
	require.NoError(t, db.Create(&oldPerm).Error)
	newPerm := domain.Permission{Name: "products.write", Group: "Inventory"}
	require.NoError(t, db.Create(&newPerm).Error)

	// Seed the role with the old permission so the diff is meaningful.
	require.NoError(t, repo.AssignPermissions(role.ID, []uint{oldPerm.ID}, 99))
	// Clear the audit row from the seed-grant so we can isolate the next write.
	require.NoError(t, db.Where("user_id = ?", 99).Delete(&domain.AuditLog{}).Error)

	require.NoError(t, repo.AssignPermissions(role.ID, []uint{newPerm.ID}, 42))

	var logs []domain.AuditLog
	require.NoError(t, db.Find(&logs).Error)
	require.Len(t, logs, 1, "exactly one audit row expected from the second grant")
	assert.Equal(t, "UPDATE", logs[0].Action)
	assert.Equal(t, "Role", logs[0].Entity)
	assert.Equal(t, strconv.FormatUint(uint64(role.ID), 10), logs[0].EntityID)
	assert.Equal(t, uint(42), logs[0].UserID)
	// The Changes JSON should record the diff: old had [products.read], new has [products.write].
	var diff struct {
		Old []string `json:"old"`
		New []string `json:"new"`
	}
	require.NoError(t, json.Unmarshal([]byte(logs[0].Changes), &diff))
	assert.ElementsMatch(t, []string{"products.read"}, diff.Old)
	assert.ElementsMatch(t, []string{"products.write"}, diff.New)
}
