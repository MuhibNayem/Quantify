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
	AutoMigrate()
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
		&domain.StockTransfer{},
		&domain.Transaction{},
		&domain.LoyaltyAccount{},
		&domain.TimeClock{},
		&domain.Job{},
	)
	if err != nil {
		logrus.Fatalf("Database schema auto-migration failed: %v", err)
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
