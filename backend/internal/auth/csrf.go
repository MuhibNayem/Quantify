package auth

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"inventory/backend/internal/repository"
)

var csrfTokenTTL = 24 * time.Hour

const csrfScanBatch = 200

// GenerateCSRFToken creates and stores a CSRF token bound to a user ID.
func GenerateCSRFToken(userID uint) (string, error) {
	token := uuid.NewString()
	cacheKey := fmt.Sprintf("csrf_token:%s", token)
	if err := repository.SetCache(cacheKey, fmt.Sprintf("%d", userID), csrfTokenTTL); err != nil {
		return "", err
	}
	return token, nil
}

// ValidateCSRFToken ensures a token exists and belongs to the authenticated user.
func ValidateCSRFToken(token string, userID uint) (bool, error) {
	if token == "" {
		return false, nil
	}
	cacheKey := fmt.Sprintf("csrf_token:%s", token)
	cachedUserID, err := repository.GetCache(cacheKey)
	if err != nil {
		return false, err
	}
	if cachedUserID == "" {
		return false, nil
	}
	value, err := strconv.ParseUint(cachedUserID, 10, 64)
	if err != nil {
		return false, err
	}
	return uint(value) == userID, nil
}

// InvalidateCSRFToken removes the provided CSRF token from cache.
func InvalidateCSRFToken(token string) error {
	if token == "" {
		return nil
	}
	cacheKey := fmt.Sprintf("csrf_token:%s", token)
	return repository.DeleteCache(cacheKey)
}

// RotateCSRFTokenFor deletes every CSRF token bound to userID, forcing the
// user to acquire a new token on the next request. Use this when a user's
// privileges (role, permissions) change so that any cached token validated
// against the old privilege set cannot be reused.
func RotateCSRFTokenFor(userID uint) error {
	client := repository.GetClient()
	if client == nil {
		return nil
	}
	ctx := context.Background()
	target := strconv.FormatUint(uint64(userID), 10)
	iter := client.Scan(ctx, 0, "csrf_token:*", csrfScanBatch).Iterator()
	keys := make([]string, 0, csrfScanBatch)
	for iter.Next(ctx) {
		key := iter.Val()
		owner, err := client.Get(ctx, key).Result()
		if err != nil {
			continue
		}
		if owner == target {
			keys = append(keys, key)
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return client.Del(ctx, keys...).Err()
}
