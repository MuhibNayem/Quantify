package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"

	"inventory/backend/internal/repository"
)

var csrfTokenTTL = 24 * time.Hour

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
