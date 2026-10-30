package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"inventory/backend/internal/bootstrap"
	"inventory/backend/internal/config"
	"inventory/backend/internal/migrations"
	"inventory/backend/internal/repository"
)

// @title           Quantify API
// @version         1.0
// @description     Inventory management and POS system API
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer token. Format: "Bearer {token}"
func main() {
	// -migrate runs the embedded SQL migrations against the database
	// pointed at by the configuration, then exits 0. Used by the docker-
	// compose `migrator` init service for explicit visibility in
	// `docker-compose ps`; the backend also auto-runs the same migrations
	// on boot, so this is a belt-and-braces guarantee for fresh
	// environments.
	migrateOnly := flag.Bool("migrate", false, "run embedded SQL migrations and exit")
	flag.Parse()

	// Initialize Logrus
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.LoadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	if *migrateOnly {
		runMigrationsOnce(cfg)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize App
	app := bootstrap.NewApp(cfg)

	// Run App
	app.Run(ctx)
}

// runMigrationsOnce opens a GORM handle, applies the embedded SQL
// migrations, and exits. Idempotent — safe to run on every deploy.
func runMigrationsOnce(cfg *config.Config) {
	// Build the DSN from the same fields the backend uses, instead of
	// requiring DATABASE_URL to be set explicitly.
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	if err := runEmbeddedMigrations(dsn); err != nil {
		logrus.Fatalf("migrations failed: %v", err)
	}
	logrus.Info("migrations complete")
}

func runEmbeddedMigrations(dsn string) error {
	// The migrations package embeds its SQL files, so we don't need a
	// separate cmd/migrator binary — this single binary handles both the
	// server and the one-shot migration mode. InitDBFromDSN fatally exits
	// on connection failure (same contract as InitDB).
	repository.InitDBFromDSN(dsn)
	return migrations.Run(repository.DB())
}
