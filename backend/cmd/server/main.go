package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"inventory/backend/internal/bootstrap"
	"inventory/backend/internal/config"
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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize Logrus
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.LoadConfig(os.Getenv("CONFIG_FILE"))
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize App
	app := bootstrap.NewApp(cfg)

	// Run App
	app.Run(ctx)
}
