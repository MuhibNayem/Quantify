package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestListNotifications_RejectsMissingUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := &NotificationHandler{}
	r.GET("/users/:id/notifications", h.ListNotifications)

	req := httptest.NewRequest("GET", "/users/1/notifications", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code, "missing user context must 401, not silently pass")
}