//go:build integration

package integration

import (
	"context"
	"database/sql"
	"testing"
	"time"

	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func SetupPostgres(t *testing.T) *gorm.DB {
	t.Helper()
	ctx := context.Background()
	pg, err := tcpostgres.RunContainer(ctx,
		tcpostgres.WithDatabase("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.WithUsername("test"),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	dsn, _ := pg.ConnectionString(ctx)
	for i := 0; i < 30; i++ {
		db, err := sql.Open("postgres", dsn)
		if err == nil && db.Ping() == nil {
			db.Close()
			break
		}
		time.Sleep(time.Second)
	}
	gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return gdb
}

func Migrate(t *testing.T, db *gorm.DB) {
	t.Helper()
	err := db.AutoMigrate(
		&domain.User{},
		&domain.Product{},
		&domain.Category{},
		&domain.SubCategory{},
		&domain.Supplier{},
		&domain.Location{},
		&domain.StockAdjustment{},
		&domain.Batch{},
		&domain.PurchaseOrder{},
		&domain.PurchaseOrderItem{},
		&domain.ReorderSuggestion{},
		&domain.DemandForecast{},
		&domain.Alert{},
		&domain.UserNotificationSettings{},
		&domain.ProductAlertSettings{},
		&domain.StockTransfer{},
		&domain.Transaction{},
		&domain.LoyaltyAccount{},
		&domain.TimeClock{},
		&domain.Job{},
		&domain.Notification{},
		&domain.AlertRoleSubscription{},
		&domain.UserNotificationSubscription{},
		&domain.GlobalSearchEntry{},
		&domain.SystemSetting{},
		&domain.Permission{},
		&domain.Role{},
		&domain.RolePermission{},
		&domain.Order{},
		&domain.OrderItem{},
		&domain.Return{},
		&domain.ReturnItem{},
		&domain.AuditLog{},
		&domain.CashDrawerSession{},
		&domain.CashDrop{},
		&domain.Promotion{},
		&domain.PurchaseReturn{},
		&domain.PurchaseReturnItem{},
	)
	if err != nil {
		t.Fatalf("auto-migrate failed: %v", err)
	}
}
