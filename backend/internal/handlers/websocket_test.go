package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestParseAuthMessage_ValidAuthReturnsToken(t *testing.T) {
	data := []byte(`{"type":"auth","token":"abc123"}`)
	token, err := parseAuthMessage(data)
	require.NoError(t, err)
	assert.Equal(t, "abc123", token)
}

func TestParseAuthMessage_RejectsWrongType(t *testing.T) {
	data := []byte(`{"type":"chat","token":"abc123"}`)
	_, err := parseAuthMessage(data)
	assert.Error(t, err)
}

func TestParseAuthMessage_RejectsMissingToken(t *testing.T) {
	data := []byte(`{"type":"auth"}`)
	_, err := parseAuthMessage(data)
	assert.Error(t, err)
}

func TestParseAuthMessage_RejectsInvalidJSON(t *testing.T) {
	data := []byte(`not-json`)
	_, err := parseAuthMessage(data)
	assert.Error(t, err)
}

func TestParseAuthMessage_RejectsEmpty(t *testing.T) {
	_, err := parseAuthMessage(nil)
	assert.Error(t, err)
	_, err = parseAuthMessage([]byte{})
	assert.Error(t, err)
}
