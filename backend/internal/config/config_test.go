package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfig_RejectsDefaultJWTSecret(t *testing.T) {
	cfg := &Config{JWTSecret: "my_secret_key", RefreshTokenSecret: "ok", MinioAccessKeyID: "ok", MinioSecretAccessKey: "ok", RabbitMQURL: "amqp://x"}
	err := validateCriticalConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestValidateConfig_RejectsDefaultMinioSecrets(t *testing.T) {
	cfg := &Config{JWTSecret: "unique", RefreshTokenSecret: "unique", MinioAccessKeyID: "minioadmin", MinioSecretAccessKey: "minioadmin", RabbitMQURL: "amqp://x"}
	err := validateCriticalConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfig_RejectsDockerComposeMinioCreds(t *testing.T) {
	// docker-compose.yml uses these defaults — operators must rotate them
	cfg := &Config{JWTSecret: "unique", RefreshTokenSecret: "unique", MinioAccessKeyID: "minioadmin4321", MinioSecretAccessKey: "minioadmin4321", RabbitMQURL: "amqp://x"}
	err := validateCriticalConfig(cfg)
	assert.Error(t, err)
}

func TestValidateConfig_AcceptsUniqueSecrets(t *testing.T) {
	cfg := &Config{JWTSecret: "u1", RefreshTokenSecret: "u2", MinioAccessKeyID: "u3", MinioSecretAccessKey: "u4", RabbitMQURL: "amqp://x"}
	assert.NoError(t, validateCriticalConfig(cfg))
}

func TestValidateConfig_RejectsEmptyJWTSecret(t *testing.T) {
	cfg := &Config{JWTSecret: "", RefreshTokenSecret: "u", MinioAccessKeyID: "u", MinioSecretAccessKey: "u", RabbitMQURL: "amqp://x"}
	err := validateCriticalConfig(cfg)
	assert.Error(t, err)
}