package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application-wide configurations.
type Config struct {
	DBHost                  string
	DBUser                  string
	DBPassword              string
	DBName                  string
	DBPort                  string
	ServerPort              string
	SMTPHost                string
	SMTPPort                int
	SMTPUser                string
	SMTPPass                string
	SMTPSender              string
	RedisClusterAddrs       []string
	RedisPassword           string
	RabbitMQURL             string
	JWTSecret               string
	RefreshTokenSecret      string
	SSLCommerzStoreID       string
	SSLCommerzStorePassword string
	SSLCommerzAPIHost       string
	SSLCommerzSuccessURL    string
	SSLCommerzFailURL       string
	SSLCommerzCancelURL     string
	SSLCommerzIPNURL        string
	BkashAPIKey             string
	BkashAPISecret          string
	BkashUsername           string
	BkashPassword           string
	BkashBaseURL            string
	MinioEndpoint           string `env:"MINIO_ENDPOINT,required"`
	MinioAccessKeyID        string `env:"MINIO_ACCESS_KEY_ID,required"`
	MinioSecretAccessKey    string `env:"MINIO_SECRET_ACCESS_KEY,required"`
	MinioBucketName         string `env:"MINIO_BUCKET_NAME,required"`
	MinioUseTLS             bool   `env:"MINIO_USE_TLS"`
	SalesTrendsCacheTTL       time.Duration
		InventoryTurnoverCacheTTL time.Duration
		ProfitMarginCacheTTL      time.Duration
		ConsumerConcurrency       int
	}
	
	// LoadConfig loads configuration from environment variables.
	func LoadConfig() *Config {
	
		dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432")) // Default to 5432 if not set
		if err != nil {
			log.Fatalf("DB_PORT must be an integer: %v", err)
		}
	
		smtpPort, err := strconv.Atoi(getEnv("SMTP_PORT", "587"))
		if err != nil {
			log.Fatalf("SMTP_PORT must be an integer: %v", err)
		}
	
		minioUseTLS, err := strconv.ParseBool(getEnv("MINIO_USE_TLS", "false"))
		if err != nil {
			log.Fatalf("MINIO_USE_TLS must be a boolean: %v", err)
		}
	
		salesTrendsCacheTTL, err := time.ParseDuration(getEnv("SALES_TRENDS_CACHE_TTL", "5m"))
		if err != nil {
			log.Fatalf("SALES_TRENDS_CACHE_TTL must be a valid duration string: %v", err)
		}
	
		inventoryTurnoverCacheTTL, err := time.ParseDuration(getEnv("INVENTORY_TURNOVER_CACHE_TTL", "5m"))
		if err != nil {
			log.Fatalf("INVENTORY_TURNOVER_CACHE_TTL must be a valid duration string: %v", err)
		}
	
		profitMarginCacheTTL, err := time.ParseDuration(getEnv("PROFIT_MARGIN_CACHE_TTL", "5m"))
		if err != nil {
			log.Fatalf("PROFIT_MARGIN_CACHE_TTL must be a valid duration string: %v", err)
		}
	
		consumerConcurrency, err := strconv.Atoi(getEnv("CONSUMER_CONCURRENCY", "1"))
		if err != nil {
			log.Fatalf("CONSUMER_CONCURRENCY must be an integer: %v", err)
		}
	
		return &Config{
			DBHost:                  getEnv("DB_HOST", "localhost"),
			DBUser:                  getEnv("DB_USER", "user"),
			DBPassword:              getEnv("DB_PASSWORD", "password"),
			DBName:                  getEnv("DB_NAME", "inventory_db"),
			DBPort:                  strconv.Itoa(dbPort),
			ServerPort:              getEnv("SERVER_PORT", "8080"),
			SMTPHost:                getEnv("SMTP_HOST", ""),
			SMTPPort:                smtpPort,
			SMTPUser:                getEnv("SMTP_USER", ""),
			SMTPPass:                getEnv("SMTP_PASS", ""),
			SMTPSender:              getEnv("SMTPSENDER", ""),
			RedisClusterAddrs:       strings.Split(getEnv("REDIS_CLUSTER_ADDRS", "redis1:6379,redis2:6379,redis3:6379"), ","),
			RedisPassword:           getEnv("REDIS_PASSWORD", "redispass"),
			RabbitMQURL:             getEnv("RABBITMQ_URL", ""),
			JWTSecret:               getEnv("JWT_SECRET", "my_secret_key"),
			RefreshTokenSecret:      getEnv("REFRESH_TOKEN_SECRET", "my_refresh_secret_key"),
			SSLCommerzStoreID:       getEnv("SSLCOMMERZ_STORE_ID", ""),	
			SSLCommerzStorePassword: getEnv("SSLCOMMERZ_STORE_PASSWORD", ""),
			SSLCommerzAPIHost:       getEnv("SSLCOMMERZ_API_HOST", "https://sandbox.sslcommerz.com"),
			SSLCommerzSuccessURL:    getEnv("SSLCOMMERZ_SUCCESS_URL", "http://localhost:8080/payment/success"),
			SSLCommerzFailURL:       getEnv("SSLCOMMERZ_FAIL_URL", "http://localhost:8080/payment/fail"),
			SSLCommerzCancelURL:     getEnv("SSLCOMMERZ_CANCEL_URL", "http://localhost:8080/payment/cancel"),
			SSLCommerzIPNURL:        getEnv("SSLCOMMERZ_IPN_URL", "http://localhost:8080/payment/ipn"),
			BkashAPIKey:             getEnv("BKASH_API_KEY", ""),
			BkashAPISecret:          getEnv("BKASH_API_SECRET", ""),
			BkashUsername:           getEnv("BKASH_USERNAME", ""),
			BkashPassword:           getEnv("BKASH_PASSWORD", ""),
			BkashBaseURL:            getEnv("BKASH_BASE_URL", "https://sandbox.bkash.com/v1.2.0-beta"),
			MinioEndpoint:           getEnv("MINIO_ENDPOINT", "localhost:9000"),
			MinioAccessKeyID:        getEnv("MINIO_ACCESS_KEY_ID", "minioadmin"),
			MinioSecretAccessKey:    getEnv("MINIO_SECRET_ACCESS_KEY", "minioadmin"),
			MinioBucketName:         getEnv("MINIO_BUCKET_NAME", "reports"),
			MinioUseTLS:             minioUseTLS,
			SalesTrendsCacheTTL:       salesTrendsCacheTTL,
			InventoryTurnoverCacheTTL: inventoryTurnoverCacheTTL,
			ProfitMarginCacheTTL:      profitMarginCacheTTL,
			ConsumerConcurrency:       consumerConcurrency,
		}
	}
	
	// getEnv gets an environment variable or returns a default value.
	func getEnv(key, defaultValue string) string {	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
