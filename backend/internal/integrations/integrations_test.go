package integrations

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWebhookRegistry_RejectsUnknown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reg := NewRegistry()
	r := gin.New()
	r.POST("/webhooks/:name", reg.HandleWebhook)

	req := httptest.NewRequest("POST", "/webhooks/unknown", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 404, w.Code)
}

func TestWebhookRegistry_AcceptsRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reg := NewRegistry()
	reg.Register(&Shopify{})
	r := gin.New()
	r.POST("/webhooks/:name", reg.HandleWebhook)

	req := httptest.NewRequest("POST", "/webhooks/shopify", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestWebhookRegistry_AcceptsQuickBooks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	reg := NewRegistry()
	reg.Register(&QuickBooks{})
	r := gin.New()
	r.POST("/webhooks/:name", reg.HandleWebhook)

	req := httptest.NewRequest("POST", "/webhooks/quickbooks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
}

func TestWebhookRegistry_RegisterReplaces(t *testing.T) {
	reg := NewRegistry()
	s1 := &Shopify{}
	reg.Register(s1)
	assert.Len(t, reg.integrations, 1)

	s2 := &Shopify{}
	reg.Register(s2)
	assert.Len(t, reg.integrations, 1)
	assert.Same(t, s2, reg.integrations["shopify"])
}