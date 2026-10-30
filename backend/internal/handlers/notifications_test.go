package handlers

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/middleware"
)

func setupNotificationSubscriptionTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&domain.AlertRoleSubscription{},
		&domain.UserNotificationSubscription{},
		&domain.Notification{},
	))

	role := domain.Role{Name: "Manager"}
	require.NoError(t, db.Create(&role).Error)

	user := domain.User{
		Username:    "alice",
		Password:    "x",
		Email:       "alice@example.com",
		PhoneNumber: "1111111111",
		RoleID:      role.ID,
	}
	require.NoError(t, db.Create(&user).Error)

	require.NoError(t, db.Create(&domain.AlertRoleSubscription{
		AlertType: "LOW_STOCK",
		Role:      "Manager",
	}).Error)

	return db
}

func TestListNotifications_RejectsMissingUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := &NotificationHandler{}
	r.GET("/users/:id/notifications", h.ListNotifications)

	req := httptest.NewRequest("GET", "/users/1/notifications", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code, "missing user context must 401, not silently pass")
}

func TestUpdateSubscription_DisablesAlertType(t *testing.T) {
	db := setupNotificationSubscriptionTestDB(t)

	var user domain.User
	require.NoError(t, db.Where("username = ?", "alice").First(&user).Error)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})
	h := &NotificationHandler{db: db}
	r.PUT("/users/:id/notifications/subscriptions/:type", h.UpdateSubscription)

	body := `{"enabled": false}`
	req := httptest.NewRequest(
		"PUT",
		fmt.Sprintf("/users/%d/notifications/subscriptions/LOW_STOCK", user.ID),
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "PUT must succeed, body=%s", w.Body.String())

	var sub domain.UserNotificationSubscription
	require.NoError(t, db.Where("user_id = ? AND type = ?", user.ID, "LOW_STOCK").First(&sub).Error)
	assert.False(t, sub.Enabled, "subscription must be disabled")
}

func TestUpdateSubscription_RejectsCrossUserWrite(t *testing.T) {
	db := setupNotificationSubscriptionTestDB(t)

	var user domain.User
	require.NoError(t, db.Where("username = ?", "alice").First(&user).Error)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Set("role", "Customer")
		c.Next()
	})
	h := &NotificationHandler{db: db}
	r.PUT("/users/:id/notifications/subscriptions/:type", h.UpdateSubscription)

	body := `{"enabled": false}`
	req := httptest.NewRequest(
		"PUT",
		fmt.Sprintf("/users/%d/notifications/subscriptions/LOW_STOCK", user.ID),
		strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code, "owner may update own subscription")

	crossReq := httptest.NewRequest(
		"PUT",
		fmt.Sprintf("/users/%d/notifications/subscriptions/LOW_STOCK", user.ID+999),
		strings.NewReader(body),
	)
	crossReq.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, crossReq)

	assert.Equal(t, 403, w2.Code, "non-owner must not update other user's subscription")
}

func TestUpdateSubscription_RejectsUnauthenticated(t *testing.T) {
	db := setupNotificationSubscriptionTestDB(t)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	h := &NotificationHandler{db: db}
	r.PUT("/users/:id/notifications/subscriptions/:type", h.UpdateSubscription)

	req := httptest.NewRequest("PUT", "/users/1/notifications/subscriptions/LOW_STOCK",
		strings.NewReader(`{"enabled": false}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 401, w.Code, "no user_id in context must 401")
}

func TestListSubscriptions_ReturnsUserPreferences(t *testing.T) {
	db := setupNotificationSubscriptionTestDB(t)

	var user domain.User
	require.NoError(t, db.Where("username = ?", "alice").First(&user).Error)

	require.NoError(t, db.Create(&domain.UserNotificationSubscription{
		UserID:  user.ID,
		Type:    "LOW_STOCK",
		Enabled: false,
	}).Error)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.ErrorHandler())
	r.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})
	h := &NotificationHandler{db: db}
	r.GET("/users/:id/notifications/subscriptions", h.ListSubscriptions)

	req := httptest.NewRequest(
		"GET",
		fmt.Sprintf("/users/%d/notifications/subscriptions", user.ID),
		nil,
	)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Contains(t, w.Body.String(), `"LOW_STOCK"`)
	assert.Contains(t, w.Body.String(), `"enabled":false`)
}
