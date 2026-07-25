package auth

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"inventory/backend/internal/repository"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	repository.RedisClient = &repository.UniversalClient{UniversalClient: rdb}
	return mr, rdb
}

func TestRotateCSRFTokenFor_InvalidatesAllTokensForUser(t *testing.T) {
	setupTestRedis(t)

	userID := uint(42)

	tokens := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		tok, err := GenerateCSRFToken(userID)
		require.NoError(t, err)
		tokens = append(tokens, tok)
	}

	for _, tok := range tokens {
		valid, err := ValidateCSRFToken(tok, userID)
		require.NoError(t, err)
		require.True(t, valid, "generated token must validate before rotation")
	}

	err := RotateCSRFTokenFor(userID)
	require.NoError(t, err)

	for _, tok := range tokens {
		valid, err := ValidateCSRFToken(tok, userID)
		assert.NoError(t, err)
		assert.False(t, valid, "old token must not validate after rotation")
	}

	newToken, err := GenerateCSRFToken(userID)
	require.NoError(t, err)
	valid, err := ValidateCSRFToken(newToken, userID)
	require.NoError(t, err)
	assert.True(t, valid, "freshly issued token must validate after rotation")
}

func TestRotateCSRFTokenFor_DoesNotAffectOtherUsers(t *testing.T) {
	setupTestRedis(t)

	userA := uint(100)
	userB := uint(200)

	tokA, err := GenerateCSRFToken(userA)
	require.NoError(t, err)
	tokB, err := GenerateCSRFToken(userB)
	require.NoError(t, err)

	err = RotateCSRFTokenFor(userA)
	require.NoError(t, err)

	validA, err := ValidateCSRFToken(tokA, userA)
	require.NoError(t, err)
	assert.False(t, validA, "rotated user's token must be invalidated")

	validB, err := ValidateCSRFToken(tokB, userB)
	require.NoError(t, err)
	assert.True(t, validB, "other user's token must remain valid after rotation")
}

func TestRotateCSRFTokenFor_NoTokensIsNoOp(t *testing.T) {
	setupTestRedis(t)

	err := RotateCSRFTokenFor(uint(999))
	assert.NoError(t, err, "rotating when no tokens exist should not error")
}
