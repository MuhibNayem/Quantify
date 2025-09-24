package repository

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"inventory/backend/internal/config"
	"inventory/backend/internal/domain" // Assuming domain package is in internal/domain
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

	// Auto-migrate the schema
	err = DB.AutoMigrate(
		&domain.Product{},
		&domain.Category{},
		&domain.SubCategory{},
		&domain.Supplier{},
		&domain.Batch{},
		&domain.StockAdjustment{},
		&domain.Alert{},
		&domain.User{},
		&domain.ProductAlertSettings{},
		&domain.UserNotificationSettings{},
		&domain.DemandForecast{},
		&domain.ReorderSuggestion{},
		&domain.PurchaseOrder{},
		&domain.PurchaseOrderItem{},
		&domain.Location{},
	)
	if err != nil {
		logrus.Fatalf("Failed to auto-migrate database schema: %v", err)
	}
	logrus.Info("Database schema auto-migrated")
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
