package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"inventory/backend/internal/config"
)

type ClusterClient struct {
	*redis.ClusterClient
}

var RedisClient *ClusterClient
var ctx = context.Background()

// InitRedis initializes the Redis client using Redis Cluster.
func InitRedis(cfg *config.Config) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        cfg.RedisClusterAddrs,
		Password:     cfg.RedisPassword,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     200,
		MinIdleConns: 20,
		PoolTimeout:  3 * time.Second,
	})

	_, err := client.Ping(ctx).Result()
	if err != nil {
		logrus.Fatalf("Could not connect to Redis Cluster: %v", err)
	}

	logrus.Info("Connected to Redis Cluster successfully!")
	RedisClient = &ClusterClient{client}
}

// CloseRedis closes the Redis client connection.
func CloseRedis() {
	if RedisClient != nil && RedisClient.ClusterClient != nil {
		if err := RedisClient.ClusterClient.Close(); err != nil {
			logrus.Errorf("Error closing Redis client: %v", err)
		}
		logrus.Info("Redis connection closed")
	}
}

// SetCache sets a key-value pair in Redis with an expiration.
func SetCache(key string, value interface{}, expiration time.Duration) error {
	return RedisClient.ClusterClient.Set(ctx, key, value, expiration).Err()
}

// GetCache gets a value from Redis.
func GetCache(key string) (string, error) {
	val, err := RedisClient.ClusterClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key does not exist
	} else if err != nil {
		logrus.Errorf("Failed to get cache for key %s: %v", key, err)
		return "", err
	}
	return val, nil
}

// DeleteCache deletes a key from Redis.
func DeleteCache(key string) error {
	return RedisClient.ClusterClient.Del(ctx, key).Err()
}
