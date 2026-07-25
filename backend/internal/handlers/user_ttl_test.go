package handlers

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"inventory/backend/internal/config"
)

func TestResolveRefreshTTL_Default30Days(t *testing.T) {
	cfg := &config.Config{}
	assert.Equal(t, 720*time.Hour, resolveRefreshTTL(cfg))
}

func TestResolveRefreshTTL_RespectsConfig(t *testing.T) {
	cfg := &config.Config{RefreshTokenTTL: 24 * time.Hour}
	assert.Equal(t, 24*time.Hour, resolveRefreshTTL(cfg))
}

func TestResolveRefreshTTL_DefaultsOnZeroOrNegative(t *testing.T) {
	cfg := &config.Config{RefreshTokenTTL: -1 * time.Hour}
	assert.Equal(t, 720*time.Hour, resolveRefreshTTL(cfg))
}