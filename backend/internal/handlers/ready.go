package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"inventory/backend/internal/migrations"
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
	// rbacReady catches the failure mode where a fresh DB was created by
	// GORM AutoMigrate but the SQL migrations never ran (the bug that left
	// Admin with an empty permissions list). Without it, the API would
	// accept traffic but every permission-gated UI button would be hidden
	// and every `RequirePermission` call beyond the Admin bypass would
	// silently 403. We'd rather flip the readiness probe red and let an
	// operator notice.
	if !h.rbacReady(ctx) {
		failed = append(failed, "rbac")
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

// rbacReady verifies that the SQL migrations have been applied AND that the
// permissions table isn't empty. The version check protects against a
// half-applied migration (DIRTY state); the permissions count check
// protects against the case where the schema exists but the RBAC seed
// data was skipped.
//
// We expect 41+ permissions (the catalogue in 0002_seed.up.sql). A
// healthy production database has every one of them.
func (h *ReadyHandler) rbacReady(ctx context.Context) bool {
	if h.db == nil {
		return false
	}
	version, dirty, err := migrations.Version(h.db)
	if err != nil {
		return false
	}
	if dirty {
		return false
	}
	if version == 0 {
		// No migrations applied at all — schema exists (thanks to
		// AutoMigrate) but reference data is missing.
		return false
	}
	var permCount int64
	if err := h.db.WithContext(ctx).
		Table("permissions").
		Count(&permCount).Error; err != nil {
		return false
	}
	return permCount > 0
}
