package main

import (
	"inventory/backend/internal/config"
	"inventory/backend/internal/migrations"
	"inventory/backend/internal/repository"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize Logrus
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	logrus.Info("Starting database migration for search index...")

	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database connection
	repository.InitDB(cfg)
	db := repository.DB
	defer repository.CloseDB()

	// Run the migration to populate the search index
	if err := migrations.PopulateSearchIndex(db); err != nil {
		logrus.Fatalf("Failed to run search index migration: %v", err)
	}

	logrus.Info("Search index migration completed successfully.")
}
