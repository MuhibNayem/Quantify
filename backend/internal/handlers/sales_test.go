package handlers

import (
	"encoding/json"
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

func setupSalesTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:sales-test-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.Order{}))
	return db
}

func seedOrders(t *testing.T, db *gorm.DB, n int) {
	t.Helper()
	orders := make([]domain.Order, n)
	for i := 0; i < n; i++ {
		orders[i] = domain.Order{
			OrderNumber:   fmt.Sprintf("ORD-%05d", i),
			TotalAmount:   float64(i + 1),
			Status:        "COMPLETED",
			PaymentMethod: "CASH",
			OrderDate:     time.Now().Add(-time.Duration(i) * time.Minute),
		}
	}
	require.NoError(t, db.Create(&orders).Error)
}

func TestListAllOrders_Pagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupSalesTestDB(t)
	seedOrders(t, db, 250)
	r := gin.New()
	h := &SalesHandler{DB: db}
	r.GET("/sales/history", h.ListAllOrders)

	req := httptest.NewRequest("GET", "/sales/history?page=2&limit=50", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)
	var resp struct {
		Data  []domain.Order `json:"data"`
		Total int            `json:"total"`
		Page  int            `json:"page"`
		Limit int            `json:"limit"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Len(t, resp.Data, 50)
	assert.Equal(t, 250, resp.Total)
	assert.Equal(t, 2, resp.Page)
	assert.Equal(t, 50, resp.Limit)
}
