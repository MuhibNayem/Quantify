package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	DBHost                    string        `koanf:"DB_HOST"`
	DBUser                    string        `koanf:"DB_USER"`
	DBPassword                string        `koanf:"DB_PASSWORD"`
	DBName                    string        `koanf:"DB_NAME"`
	DBPort                    string        `koanf:"DB_PORT"`
	ServerPort                string        `koanf:"SERVER_PORT"`
	WSAllowedOrigins          []string      `koanf:"WS_ALLOWED_ORIGINS"`
	CORSAllowedOrigins        []string      `koanf:"CORS_ALLOWED_ORIGINS"`
	SMTPHost                  string        `koanf:"SMTP_HOST"`
	SMTPPort                  int           `koanf:"SMTP_PORT"`
	SMTPUser                  string        `koanf:"SMTP_USER"`
	SMTPPass                  string        `koanf:"SMTP_PASS"`
	SMTPSender                string        `koanf:"SMTPSENDER"`
	RedisAddr                 string        `koanf:"REDIS_ADDR"`
	RedisClusterAddrs         []string      `koanf:"REDIS_CLUSTER_ADDRS"`
	RedisPassword             string        `koanf:"REDIS_PASSWORD"`
	RabbitMQURL               string        `koanf:"RABBITMQ_URL"`
	JWTSecret                 string        `koanf:"JWT_SECRET"`
	RefreshTokenSecret        string        `koanf:"REFRESH_TOKEN_SECRET"`
	SSLCommerzStoreID         string        `koanf:"SSLCOMMERZ_STORE_ID"`
	SSLCommerzStorePassword   string        `koanf:"SSLCOMMERZ_STORE_PASSWORD"`
	SSLCommerzAPIHost         string        `koanf:"SSLCOMMERZ_API_HOST"`
	SSLCommerzSuccessURL      string        `koanf:"SSLCOMMERZ_SUCCESS_URL"`
	SSLCommerzFailURL         string        `koanf:"SSLCOMMERZ_FAIL_URL"`
	SSLCommerzCancelURL       string        `koanf:"SSLCOMMERZ_CANCEL_URL"`
	SSLCommerzIPNURL          string        `koanf:"SSLCOMMERZ_IPN_URL"`
	BkashAPIKey               string        `koanf:"BKASH_API_KEY"`
	BkashAPISecret            string        `koanf:"BKASH_API_SECRET"`
	BkashUsername             string        `koanf:"BKASH_USERNAME"`
	BkashPassword             string        `koanf:"BKASH_PASSWORD"`
	BkashBaseURL              string        `koanf:"BKASH_BASE_URL"`
	MinioEndpoint             string        `koanf:"MINIO_ENDPOINT"`
	MinioAccessKeyID          string        `koanf:"MINIO_ACCESS_KEY_ID"`
	MinioSecretAccessKey      string        `koanf:"MINIO_SECRET_ACCESS_KEY"`
	MinioBucketName           string        `koanf:"MINIO_BUCKET_NAME"`
	MinioUseTLS               bool          `koanf:"MINIO_USE_TLS"`
	SalesTrendsCacheTTL       time.Duration `koanf:"SALES_TRENDS_CACHE_TTL"`
	InventoryTurnoverCacheTTL time.Duration `koanf:"INVENTORY_TURNOVER_CACHE_TTL"`
	ProfitMarginCacheTTL      time.Duration `koanf:"PROFIT_MARGIN_CACHE_TTL"`
	ConsumerConcurrency       int           `koanf:"CONSUMER_CONCURRENCY"`
	AIServiceURL              string        `koanf:"AI_SERVICE_URL"`
	RefreshTokenTTL           time.Duration `koanf:"REFRESH_TOKEN_TTL"`
	AppEnv                    string        `koanf:"APP_ENV"`
	OTLPEndpoint              string        `koanf:"OTLP_ENDPOINT"`
	ServiceName               string        `koanf:"SERVICE_NAME"`
	RateLimitRequests         int           `koanf:"RATE_LIMIT_REQUESTS"`
	RateLimitWindow           time.Duration `koanf:"RATE_LIMIT_WINDOW"`
	AuthRateLimitRequests     int           `koanf:"AUTH_RATE_LIMIT_REQUESTS"`
	AuthRateLimitWindow       time.Duration `koanf:"AUTH_RATE_LIMIT_WINDOW"`
	TrustedProxies            []string      `koanf:"TRUSTED_PROXIES"`
}

func LoadConfig(path string) (*Config, error) {
	k := koanf.New(".")

	_ = k.Load(confmap.Provider(defaults(), "."), nil)

	if path != "" {
		_ = k.Load(file.Provider(path), toml.Parser())
	}

	_ = k.Load(env.Provider("", ".", nil), nil)

	var cfg Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			Result:           &cfg,
			WeaklyTypedInput: true,
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
			),
		},
	}); err != nil {
		return nil, err
	}

	cfg.WSAllowedOrigins = trimSlice(cfg.WSAllowedOrigins)
	cfg.CORSAllowedOrigins = trimSlice(cfg.CORSAllowedOrigins)
	cfg.RedisClusterAddrs = trimSlice(cfg.RedisClusterAddrs)
	cfg.TrustedProxies = trimSlice(cfg.TrustedProxies)

	if err := validateCriticalConfig(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	return &cfg, nil
}

func defaults() map[string]any {
	return map[string]any{
		"DB_HOST":                      "localhost",
		"DB_USER":                      "user",
		"DB_PASSWORD":                  "password",
		"DB_NAME":                      "inventory_db",
		"DB_PORT":                      "5432",
		"SERVER_PORT":                  "8080",
		"WS_ALLOWED_ORIGINS":           "http://localhost:5173,http://localhost:3000,http://127.0.0.1:5173,http://127.0.0.1:3000",
		"CORS_ALLOWED_ORIGINS":         "http://localhost:5173,http://localhost:3000,http://127.0.0.1:5173,http://127.0.0.1:3000",
		"SMTP_HOST":                    "",
		"SMTP_PORT":                    587,
		"SMTP_USER":                    "",
		"SMTP_PASS":                    "",
		"SMTPSENDER":                   "",
		"REDIS_ADDR":                   "",
		"REDIS_CLUSTER_ADDRS":          "redis1:6379,redis2:6379,redis3:6379",
		"REDIS_PASSWORD":               "redispass",
		"RABBITMQ_URL":                 "",
		"JWT_SECRET":                   "my_secret_key",
		"REFRESH_TOKEN_SECRET":         "my_refresh_secret_key",
		"SSLCOMMERZ_STORE_ID":          "",
		"SSLCOMMERZ_STORE_PASSWORD":    "",
		"SSLCOMMERZ_API_HOST":          "https://sandbox.sslcommerz.com",
		"SSLCOMMERZ_SUCCESS_URL":       "http://localhost:8080/payment/success",
		"SSLCOMMERZ_FAIL_URL":          "http://localhost:8080/payment/fail",
		"SSLCOMMERZ_CANCEL_URL":        "http://localhost:8080/payment/cancel",
		"SSLCOMMERZ_IPN_URL":           "http://localhost:8080/payment/ipn",
		"BKASH_API_KEY":                "",
		"BKASH_API_SECRET":             "",
		"BKASH_USERNAME":               "",
		"BKASH_PASSWORD":               "",
		"BKASH_BASE_URL":               "https://sandbox.bkash.com/v1.2.0-beta",
		"MINIO_ENDPOINT":               "localhost:9000",
		"MINIO_ACCESS_KEY_ID":          "minioadmin",
		"MINIO_SECRET_ACCESS_KEY":      "minioadmin",
		"MINIO_BUCKET_NAME":            "reports",
		"MINIO_USE_TLS":                false,
		"SALES_TRENDS_CACHE_TTL":       "5m",
		"INVENTORY_TURNOVER_CACHE_TTL": "5m",
		"PROFIT_MARGIN_CACHE_TTL":      "5m",
		"CONSUMER_CONCURRENCY":         1,
		"AI_SERVICE_URL":               "http://localhost:8001",
		"REFRESH_TOKEN_TTL":            "720h",
		"APP_ENV":                      "dev",
		"OTLP_ENDPOINT":                "",
		"SERVICE_NAME":                 "quantify-backend",
		"RATE_LIMIT_REQUESTS":          300,
		"RATE_LIMIT_WINDOW":            "1m",
		"AUTH_RATE_LIMIT_REQUESTS":     10,
		"AUTH_RATE_LIMIT_WINDOW":       "1m",
		"TRUSTED_PROXIES":              "",
	}
}

func trimSlice(s []string) []string {
	out := make([]string, 0, len(s))
	for _, v := range s {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func validateCriticalConfig(cfg *Config) error {
	if cfg.RabbitMQURL == "" {
		return fmt.Errorf("RABBITMQ_URL is required")
	}
	if strings.TrimSpace(cfg.SSLCommerzStorePassword) == "" {
		return fmt.Errorf("SSLCOMMERZ_STORE_PASSWORD is required (set in env or .env)")
	}

	disallowedSecrets := map[string][]string{
		"JWT_SECRET":              {"my_secret_key", "secret", "changeme", "password", "12345", ""},
		"REFRESH_TOKEN_SECRET":    {"my_refresh_secret_key", "secret", "changeme", "password", "12345", ""},
		"MINIO_ACCESS_KEY_ID":     {"minioadmin", "minioadmin4321"},
		"MINIO_SECRET_ACCESS_KEY": {"minioadmin", "minioadmin4321"},
	}

	for field, badValues := range disallowedSecrets {
		actual := strings.TrimSpace(secretValueFor(cfg, field))
		for _, bad := range badValues {
			if actual == bad {
				return fmt.Errorf("%s must not use default value %q (set a strong unique value)", field, bad)
			}
		}
	}
	return nil
}

func secretValueFor(c *Config, field string) string {
	switch field {
	case "JWT_SECRET":
		return c.JWTSecret
	case "REFRESH_TOKEN_SECRET":
		return c.RefreshTokenSecret
	case "MINIO_ACCESS_KEY_ID":
		return c.MinioAccessKeyID
	case "MINIO_SECRET_ACCESS_KEY":
		return c.MinioSecretAccessKey
	}
	return ""
}
