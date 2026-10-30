package auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// AccessTokenCookie holds the short-lived JWT. HttpOnly: JS can never read it.
	AccessTokenCookie = "access_token"
	// RefreshTokenCookie holds the rotating refresh token, scoped to the auth
	// endpoints so it is not sent on ordinary API calls.
	RefreshTokenCookie = "refresh_token"
	// CSRFTokenCookie is intentionally readable by the SPA so it can echo the
	// value in the X-CSRF-Token header (double-submit pattern).
	CSRFTokenCookie = "csrf_token"

	// AccessTokenTTL is the lifetime of an access token / session cookie.
	AccessTokenTTL = 8 * time.Hour
)

// SetSessionCookies writes the access, refresh, and CSRF cookies. `secure`
// must be true in production (HTTPS); SameSite=Lax is correct because the SPA
// and API share a site (subdomains/ports do not matter).
func SetSessionCookies(c *gin.Context, access, refresh, csrf string, refreshTTL time.Duration, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(AccessTokenCookie, access, int(AccessTokenTTL.Seconds()), "/", "", secure, true)
	c.SetCookie(RefreshTokenCookie, refresh, int(refreshTTL.Seconds()), "/api/v1/users", "", secure, true)
	SetCSRFCookie(c, csrf, secure)
}

// SetCSRFCookie writes the non-HttpOnly double-submit token. The SPA mirrors
// it in an X-CSRF-Token header; it is also returned in auth responses for
// deployments where the API and SPA are on different subdomains.
func SetCSRFCookie(c *gin.Context, csrf string, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(CSRFTokenCookie, csrf, int(AccessTokenTTL.Seconds()), "/", "", secure, false)
}

// ClearSessionCookies expires all three session cookies.
func ClearSessionCookies(c *gin.Context, secure bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(AccessTokenCookie, "", -1, "/", "", secure, true)
	c.SetCookie(RefreshTokenCookie, "", -1, "/api/v1/users", "", secure, true)
	c.SetCookie(CSRFTokenCookie, "", -1, "/", "", secure, false)
}

// AccessTokenFromRequest returns the JWT from the Authorization header (API
// clients) or the httpOnly access cookie (browser sessions).
func AccessTokenFromRequest(c *gin.Context) string {
	if header := c.GetHeader("Authorization"); header != "" {
		if token := strings.TrimPrefix(header, "Bearer "); token != header {
			return token
		}
	}
	if cookie, err := c.Cookie(AccessTokenCookie); err == nil {
		return cookie
	}
	return ""
}

// RefreshTokenFromRequest prefers an explicit body token (API clients) and
// falls back to the httpOnly refresh cookie.
func RefreshTokenFromRequest(c *gin.Context, bodyToken string) string {
	if bodyToken != "" {
		return bodyToken
	}
	if cookie, err := c.Cookie(RefreshTokenCookie); err == nil {
		return cookie
	}
	return ""
}

// CSRFTokenFromRequest returns the double-submit token from the header (or the
// cookie as a last resort for non-browser clients that mirror cookies).
func CSRFTokenFromRequest(c *gin.Context) string {
	if header := c.GetHeader("X-CSRF-Token"); header != "" {
		return header
	}
	if cookie, err := c.Cookie(CSRFTokenCookie); err == nil {
		return cookie
	}
	return ""
}

// SecureCookies returns whether cookies must carry the Secure flag for the
// given environment.
func SecureCookies(appEnv string) bool {
	env := strings.ToLower(strings.TrimSpace(appEnv))
	return env != "" && env != "dev" && env != "development" && env != "test"
}
