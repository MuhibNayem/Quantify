package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildOriginChecker_AllowsListedOrigin(t *testing.T) {
	check := buildOriginChecker([]string{"https://app.example.com"})
	assert.True(t, check(reqWithOrigin("https://app.example.com")))
	assert.False(t, check(reqWithOrigin("https://evil.example.com")))
	assert.False(t, check(reqWithOrigin("")))
}

func reqWithOrigin(origin string) *http.Request {
	r := httptest.NewRequest("GET", "/ws", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	return r
}
