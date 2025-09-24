package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// Config holds all application-wide configurations.
type Config struct {
	DBHost             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBPort             string
	ServerPort         string
	SMTPHost           string
	SMTPPort           int
	SMTPUser           string
	SMTPPass           string
	SMTPSender         string
	RedisClusterAddrs  []string
	RedisPassword      string
	RabbitMQURL        string
	JWTSecret          string
	RefreshTokenSecret string
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

	return &Config{
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBUser:             getEnv("DB_USER", "user"),
		DBPassword:         getEnv("DB_PASSWORD", "password"),
		DBName:             getEnv("DB_NAME", "inventory_db"),
		DBPort:             strconv.Itoa(dbPort),
		ServerPort:         getEnv("SERVER_PORT", "8080"),
		SMTPHost:           getEnv("SMTP_HOST", ""),
		SMTPPort:           smtpPort,
		SMTPUser:           getEnv("SMTP_USER", ""),
		SMTPPass:           getEnv("SMTP_PASS", ""),
		SMTPSender:         getEnv("SMTPSENDER", ""),
		RedisClusterAddrs:  strings.Split(getEnv("REDIS_CLUSTER_ADDRS", "redis1:6379,redis2:6379,redis3:6379"), ","),
		RedisPassword:      getEnv("REDIS_PASSWORD", "redispass"),
		RabbitMQURL:        getEnv("RABBITMQ_URL", ""),
		JWTSecret:          getEnv("JWT_SECRET", "my_secret_key"),
		RefreshTokenSecret: getEnv("REFRESH_TOKEN_SECRET", "my_refresh_secret_key"),
	}
}

// getEnv gets an environment variable or returns a default value.
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
