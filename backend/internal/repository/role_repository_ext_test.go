package repository

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCacheKey = "rbac:role:Manager:permissions"

func newTestRoleRepo(t *testing.T) (*roleRepository, *miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { rdb.Close() })
	return &roleRepository{redis: rdb}, mr, rdb
}

func TestPermissionCache_InvalidatedOnGrant(t *testing.T) {
	repo, _, rdb := newTestRoleRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.SetPermissionsCache(ctx, "Manager", []string{"products.read"}))
	exists, _ := rdb.Exists(ctx, testCacheKey).Result()
	require.Equal(t, int64(1), exists, "cache must be set before invalidation")

	require.NoError(t, repo.InvalidateRoleCache(ctx, "Manager"))

	exists, _ = rdb.Exists(ctx, testCacheKey).Result()
	assert.Equal(t, int64(0), exists, "cache must be cleared immediately on invalidation")
}

func TestPermissionCache_TTLIsShort(t *testing.T) {
	repo, mr, _ := newTestRoleRepo(t)
	ctx := context.Background()
	require.NoError(t, repo.SetPermissionsCache(ctx, "Manager", []string{"products.read"}))

	ttl := mr.TTL(testCacheKey)
	assert.Positive(t, ttl, "TTL must be positive")
	assert.LessOrEqual(t, ttl, 60*time.Second, "TTL must be <= 60s")
}

func TestPermissionCache_TTLOnReadBack(t *testing.T) {
	repo, _, rdb := newTestRoleRepo(t)
	ctx := context.Background()

	require.NoError(t, repo.SetPermissionsCache(ctx, "Manager", []string{"products.read"}))
	ttl, err := rdb.TTL(ctx, testCacheKey).Result()
	require.NoError(t, err)
	assert.LessOrEqual(t, ttl, 60*time.Second, "TTL via redis client must be <= 60s")
	assert.Positive(t, ttl, "TTL via redis client must be positive")
}
