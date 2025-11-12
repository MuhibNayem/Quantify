package repository

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"gorm.io/driver/postgres"

	"gorm.io/gorm"

	"inventory/backend/internal/config"

	"inventory/backend/internal/domain"
)

var DB *gorm.DB

func InitDB(cfg *config.Config) {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",

		cfg.DBHost,

		cfg.DBUser,

		cfg.DBPassword,

		cfg.DBName,

		cfg.DBPort,
	)

	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {

		logrus.Fatalf("Failed to connect to database: %v", err)

	}

	logrus.Info("Database connection established")

	// Temporarily Auto-migrate the schema to apply the unique index

	AutoMigrate()
	ensureBarcodeUniqueIndex()

	ensureDefaultAlertSubscriptions()

}

func AutoMigrate() {

	err := DB.AutoMigrate(

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

		&domain.GlobalSearchEntry{},
	)

	if err != nil {

		logrus.Fatalf("Database schema auto-migration failed: %v", err)

	}

	logrus.Info("Database schema auto-migrated")

}

func ensureDefaultAlertSubscriptions() {

	defaultTypes := []string{"LOW_STOCK", "OUT_OF_STOCK", "OVERSTOCK", "EXPIRY_ALERT"}

	for _, alertType := range defaultTypes {

		var count int64

		if err := DB.Model(&domain.AlertRoleSubscription{}).
			Where("alert_type = ? AND role = ?", alertType, "Admin").
			Count(&count).Error; err != nil {

			logrus.Errorf("Failed to check default alert subscription for %s: %v", alertType, err)

			continue

		}

		if count == 0 {

			subscription := domain.AlertRoleSubscription{

				AlertType: alertType,

				Role: "Admin",
			}

			if err := DB.Create(&subscription).Error; err != nil {

				logrus.Errorf("Failed to seed default alert subscription for %s: %v", alertType, err)

			} else {

				logrus.Infof("Seeded default alert subscription %s -> Admin", alertType)

			}

		}

	}

}

func ensureBarcodeUniqueIndex() {
	if DB == nil {
		return
	}

	dropConstraint := `
DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM pg_constraint
		WHERE conname = 'products_barcode_upc_key'
	) THEN
		ALTER TABLE products DROP CONSTRAINT products_barcode_upc_key;
	END IF;
END $$;
`
	if err := DB.Exec(dropConstraint).Error; err != nil {
		logrus.Errorf("Failed to drop legacy barcode constraint: %v", err)
	}

	createPartialIndex := `
CREATE UNIQUE INDEX IF NOT EXISTS idx_products_barcode_upc_not_null
ON products (barcode_upc)
WHERE barcode_upc IS NOT NULL AND barcode_upc <> '';
`
	if err := DB.Exec(createPartialIndex).Error; err != nil {
		logrus.Errorf("Failed to ensure partial barcode index: %v", err)
	}
}

// CloseDB closes the database connection.

func CloseDB() {

	sqlDB, err := DB.DB()

	if err != nil {

		logrus.Errorf("Error getting underlying DB: %v", err)

		return

	}

	if err := sqlDB.Close(); err != nil {

		logrus.Errorf("Error closing database connection: %v", err)

	}

	logrus.Info("Database connection closed")

}
