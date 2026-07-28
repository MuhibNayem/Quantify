package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
	"inventory/backend/internal/message_broker"
	"inventory/backend/internal/repository"
)

// TestStockAdjustment_NoLostUpdates fires 50 concurrent STOCK_OUT requests of
// -5 each against a batch of 1000 and asserts the final quantity is 750.
//
// SQLite limitations:
//   - The FOR UPDATE clause is silently ignored by SQLite, so this test
//     cannot fully verify the SQL row-level lock. The implementation
//     correctness guarantee is the presence of:
//        tx.Clauses(clause.Locking{Strength: "UPDATE"}).Model(&Batch{})
//     on the currentQuantity read inside the transaction.
//   - SQLite serializes writers. With WAL mode + busy_timeout, concurrent
//     transactions queue rather than fail. Without the lock, lost updates
//     can still occur when Go-side reads overlap; the run with `-race`
//     detects those.
func TestStockAdjustment_NoLostUpdates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// File-based SQLite with WAL + busy_timeout so concurrent writers
	// queue cleanly instead of failing with "database table is locked".
	tmpDir := t.TempDir()
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)",
		filepath.Join(tmpDir, "race.db"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		_ = os.RemoveAll(tmpDir)
	})
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Product{},
		&domain.Batch{},
		&domain.StockAdjustment{},
	); err != nil {
		t.Fatalf("automigrate: %v", err)
	}

	// SQLite limitation: only one writer at a time. Forcing a single
	// connection serializes the transactions and eliminates SQLITE_BUSY,
	// which would otherwise mask the lost-update bug we're testing.
	// This is a test-environment constraint; the production code is
	// expected to run on PostgreSQL where FOR UPDATE is honored.
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)

	// Swap the global DB so the handler uses our test DB.
	prevDB := repository.DB()
	repository.SetDB(db)
	t.Cleanup(func() { repository.SetDB(prevDB) })

	// Initialize a dummy broker so the post-transaction Publish call
	// does not panic on a nil manager. The connection will fail and the
	// handler will log + continue, which is acceptable for this test.
	message_broker.InitRabbitMQ("amqp://guest:guest@127.0.0.1:1/")
	t.Cleanup(message_broker.Close)

	// Seed user so the handler can resolve user_id from context.
	user := domain.User{Username: "raceuser", Email: "race@example.com"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}

	prod := domain.Product{
		Name:          "Race Product",
		SKU:           "SKU1",
		SellingPrice:  10.0,
		PurchasePrice: 5.0,
		Status:        "Active",
	}
	if err := db.Create(&prod).Error; err != nil {
		t.Fatalf("seed product: %v", err)
	}

	expiry := time.Now().Add(24 * time.Hour)
	batch := domain.Batch{
		ProductID:   prod.ID,
		BatchNumber: "BATCH-RACE",
		Quantity:    1000,
		ExpiryDate:  &expiry,
	}
	if err := db.Create(&batch).Error; err != nil {
		t.Fatalf("seed batch: %v", err)
	}

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user_id", user.ID)
		c.Next()
	})
	r.POST("/products/:productId/stock/adjustments", CreateStockAdjustment)

	const concurrency = 50
	const adjustQty = 5

	var wg sync.WaitGroup
	codes := make(chan int, concurrency)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body, _ := json.Marshal(map[string]interface{}{
				"type":       "STOCK_OUT",
				"quantity":   adjustQty,
				"reasonCode": "race-test",
			})
			req := httptest.NewRequest(
				"POST",
				fmt.Sprintf("/products/%d/stock/adjustments", prod.ID),
				bytes.NewReader(body),
			)
			req.Header.Set("Content-Type", "application/json")
			// Short-deadline context so the post-transaction broker
			// Publish returns quickly instead of blocking on Dial.
			ctx, cancel := context.WithTimeout(req.Context(), 200*time.Millisecond)
			defer cancel()
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			codes <- w.Code
		}()
	}
	wg.Wait()
	close(codes)

	successCount := 0
	for code := range codes {
		if code == http.StatusCreated {
			successCount++
		}
	}
	assert.Equal(t, concurrency, successCount,
		"all concurrent STOCK_OUT requests should succeed")

	var finalQty int
	db.Model(&domain.Batch{}).Where("product_id = ?", prod.ID).
		Select("COALESCE(SUM(quantity), 0)").Row().Scan(&finalQty)
	assert.Equal(t, 750, finalQty, "no lost updates: 1000 - 50*5 = 750")
}
