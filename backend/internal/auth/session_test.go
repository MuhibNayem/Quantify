package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetSessionCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	SetSessionCookies(c, "access-jwt", "refresh-token", "csrf-token", time.Hour, true)

	cookies := w.Result().Cookies()
	byName := make(map[string]*http.Cookie, len(cookies))
	for _, cookie := range cookies {
		byName[cookie.Name] = cookie
	}

	access := byName[AccessTokenCookie]
	assert.NotNil(t, access)
	assert.True(t, access.HttpOnly, "access token must be HttpOnly")
	assert.True(t, access.Secure, "access token must be Secure outside dev")
	assert.Equal(t, http.SameSiteLaxMode, access.SameSite)
	assert.Equal(t, "/", access.Path)

	refresh := byName[RefreshTokenCookie]
	assert.NotNil(t, refresh)
	assert.True(t, refresh.HttpOnly, "refresh token must be HttpOnly")
	assert.Equal(t, "/api/v1/users", refresh.Path, "refresh cookie must be scoped to auth endpoints")

	csrf := byName[CSRFTokenCookie]
	assert.NotNil(t, csrf)
	assert.False(t, csrf.HttpOnly, "CSRF token must be readable by the SPA")
	assert.Equal(t, "csrf-token", csrf.Value)
}

func TestSecureCookies(t *testing.T) {
	assert.False(t, SecureCookies("dev"))
	assert.False(t, SecureCookies("development"))
	assert.False(t, SecureCookies("test"))
	assert.True(t, SecureCookies("staging"))
	assert.True(t, SecureCookies("production"))
}
