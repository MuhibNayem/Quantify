package handlers

import (
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:crm-test-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.User{}))
	return db
}

func seedUser(t *testing.T, db *gorm.DB, username, email, phone string) domain.User {
	t.Helper()
	user := domain.User{
		Username:    username,
		Email:       email,
		PhoneNumber: phone,
		Password:    "secret-hash",
	}
	require.NoError(t, db.Create(&user).Error)
	return user
}

func TestGetCustomer_Uniform404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	seedUser(t, db, "alice", "a@x.com", "555-1234")

	r := gin.New()
	h := &CRMHandler{db: db}
	r.GET("/crm/customers/:identifier", h.GetCustomer)

	for _, id := range []string{"555-9999", "missing@x.com", "ghost", "9999"} {
		req := httptest.NewRequest("GET", "/crm/customers/"+id, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, 404, w.Code, "missing identifier %s should be 404", id)
		assert.Contains(t, w.Body.String(), "not found")
	}
}

func TestGetCustomer_ValidLookupsStillWork(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	alice := seedUser(t, db, "alice", "a@x.com", "555-1234")

	r := gin.New()
	h := &CRMHandler{db: db}
	r.GET("/crm/customers/:identifier", h.GetCustomer)

	for _, id := range []string{
		fmt.Sprintf("%d", alice.ID),
		"a@x.com",
		"555-1234",
		"alice",
	} {
		req := httptest.NewRequest("GET", "/crm/customers/"+id, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code, "identifier %s should be 200", id)
		assert.NotContains(t, w.Body.String(), alice.Password)
	}
}
