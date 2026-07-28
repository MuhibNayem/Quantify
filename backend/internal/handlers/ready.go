package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"inventory/backend/internal/storage"
)

// ReadyHandler reports whether the process and its external dependencies are
// ready to serve traffic. Used as a Kubernetes readiness probe, complementing
// the liveness-only /health endpoint.
type ReadyHandler struct {
	db       *gorm.DB
	redis    redis.UniversalClient
	amqp     *amqp.Connection
	uploader storage.Uploader
}

func NewReadyHandler(db *gorm.DB, rdb redis.UniversalClient, conn *amqp.Connection, uploader storage.Uploader) *ReadyHandler {
	return &ReadyHandler{db: db, redis: rdb, amqp: conn, uploader: uploader}
}

// Ready godoc
// @Summary Readiness probe
// @Description Reports whether the server and its dependencies (DB, Redis, AMQP, MinIO) are ready
// @Tags health
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /ready [get]
func (h *ReadyHandler) Ready(c *gin.Context) {
	ctx := c.Request.Context()
	failed := []string{}

	if !h.dbReady(ctx) {
		failed = append(failed, "db")
	}
	if !h.redisReady(ctx) {
		failed = append(failed, "redis")
	}
	if !h.amqpReady() {
		failed = append(failed, "amqp")
	}
	if !h.minioReady(ctx) {
		failed = append(failed, "minio")
	}

	if len(failed) > 0 {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "NOT_READY", "failed": failed})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "READY"})
}

func (h *ReadyHandler) dbReady(ctx context.Context) bool {
	if h.db == nil {
		return false
	}
	sqlDB, err := h.db.DB()
	if err != nil {
		return false
	}
	return sqlDB.PingContext(ctx) == nil
}

func (h *ReadyHandler) redisReady(ctx context.Context) bool {
	if h.redis == nil {
		return false
	}
	return h.redis.Ping(ctx).Err() == nil
}

func (h *ReadyHandler) amqpReady() bool {
	return h.amqp != nil && !h.amqp.IsClosed()
}

func (h *ReadyHandler) minioReady(ctx context.Context) bool {
	if h.uploader == nil {
		return false
	}
	exists, err := h.uploader.BucketExists(ctx, "reports")
	return err == nil && exists
}
