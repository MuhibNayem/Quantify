package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/config"
	"inventory/backend/internal/domain"
	"inventory/backend/internal/handlers"
	"inventory/backend/internal/repository"
	"inventory/backend/internal/requests"
	"inventory/backend/internal/services"
	"inventory/backend/internal/websocket"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(
		&domain.User{},
		&domain.Product{},
		&domain.Batch{},
		&domain.StockAdjustment{},
		&domain.Transaction{},
		&domain.LoyaltyAccount{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.Return{},
		&domain.ReturnItem{},
		&domain.SystemSetting{},
		&domain.Role{},
		&domain.Permission{},
		&domain.RolePermission{},
	)
	return db
}

func TestRefundReturnFlow(t *testing.T) {
	db := setupTestDB()
	settingsRepo := repository.NewSettingsRepository(db)
	settingsService := services.NewSettingsService(settingsRepo)

	salesHandler := handlers.NewSalesHandler(db, settingsService, nil)

	hub := websocket.NewHub()
	go hub.Run()

	cfg := &config.Config{}
	// Create handler
	notificationRepo := repository.NewNotificationRepository(db)
	returnHandler := handlers.NewReturnHandler(db, cfg, settingsService, hub, notificationRepo, nil)

	// Setup Router
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", uint(1)) // Mock Auth
		c.Next()
	})
	r.POST("/sales/checkout", salesHandler.Checkout)
	r.POST("/returns/request", returnHandler.RequestReturn)
	r.POST("/returns/:id/process", returnHandler.ProcessReturn)

	// Seed Data
	product := domain.Product{
		Name:          "Test Product",
		SKU:           "TEST-SKU",
		SellingPrice:  100.0,
		PurchasePrice: 50.0,
		Status:        "Active",
	}
	db.Create(&product)

	batch := domain.Batch{
		ProductID:   product.ID,
		BatchNumber: "BATCH-1",
		Quantity:    10,
	}
	db.Create(&batch)

	user := domain.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	db.Create(&user)

	// 1. Checkout
	checkoutReq := requests.CheckoutRequest{
		Items: []requests.CheckoutItem{
			{ProductID: product.ID, Quantity: 2},
		},
		PaymentMethod: "cash",
		CustomerID:    &user.ID,
	}
	body, _ := json.Marshal(checkoutReq)
	req, _ := http.NewRequest("POST", "/sales/checkout", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify Order Created
	var order domain.Order
	db.Preload("OrderItems").First(&order)
	assert.Equal(t, 1, len(order.OrderItems))
	assert.Equal(t, 2, order.OrderItems[0].Quantity)
	assert.Equal(t, "COMPLETED", order.Status)

	// 2. Request Return
	returnReq := map[string]interface{}{
		"order_number": order.OrderNumber,
		"items": []map[string]interface{}{
			{
				"order_item_id": order.OrderItems[0].ID,
				"quantity":      1,
				"condition":     "GOOD",
				"reason":        "Defective",
			},
		},
	}
	body, _ = json.Marshal(returnReq)
	req, _ = http.NewRequest("POST", "/returns/request", bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify Return Created
	var returnRecord domain.Return
	db.First(&returnRecord)
	assert.Equal(t, "PENDING", returnRecord.Status)
	assert.Equal(t, 100.0, returnRecord.RefundAmount) // 1 * 100.0

	// 3. Process Return (Approve)
	processReq := map[string]string{"action": "approve"}
	body, _ = json.Marshal(processReq)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/returns/%d/process", returnRecord.ID), bytes.NewBuffer(body))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify Stock Adjustment (Stock In)
	var adjustment domain.StockAdjustment
	db.Where("type = ? AND reason_code = ?", "STOCK_IN", "RETURN").First(&adjustment)
	assert.Equal(t, 1, adjustment.Quantity)

	// Verify Refund Transaction
	var transaction domain.Transaction
	db.Where("status = ?", "REFUNDED").First(&transaction)
	assert.Equal(t, int64(10000), transaction.Amount) // 100.0 * 100

	// Verify Order Item Returned Qty
	var orderItem domain.OrderItem
	db.First(&orderItem, order.OrderItems[0].ID)
	assert.Equal(t, 1, orderItem.ReturnedQty)
	assert.True(t, orderItem.IsReturned)
}
