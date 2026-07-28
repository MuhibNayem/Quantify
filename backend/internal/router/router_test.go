package router

import (
	"net/http/httptest"
	"testing"

	"inventory/backend/internal/app"
	"inventory/backend/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORS_UsesConfigOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{CORSAllowedOrigins: []string{"https://app.example.com"}}
	c := &app.Container{Cfg: cfg}
	r := SetupRouter(c)

	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set("Origin", "https://app.example.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
}
