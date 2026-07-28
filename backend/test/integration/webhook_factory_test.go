//go:build integration

package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"inventory/backend/internal/integrations"
)

func TestWebhookRegistryDispatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	registry := integrations.NewRegistry()
	registry.Register(&integrations.Shopify{})

	router := gin.New()
	router.POST("/webhooks/:name", registry.HandleWebhook)

	t.Run("known integration returns 200", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/webhooks/shopify", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unknown integration returns 404", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/webhooks/unknown", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestWebhookRegistryRegisterAndReplace(t *testing.T) {
	registry := integrations.NewRegistry()

	registry.Register(&integrations.Shopify{})
	registry.Register(&integrations.Shopify{})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/webhooks/:name", registry.HandleWebhook)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/webhooks/shopify", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}
