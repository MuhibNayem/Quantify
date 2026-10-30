package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	appErrors "inventory/backend/internal/errors"
	"inventory/backend/internal/repository"
)

// RateLimit enforces a fixed-window request budget per client IP using Redis.
// It fails open if Redis is unavailable: availability of the API must not
// depend on the rate-limit store, and the ingress/edge limiter remains a
// second line of defense.
func RateLimit(limit int, window time.Duration, scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limit <= 0 || window <= 0 {
			c.Next()
			return
		}

		client := repository.GetClient()
		if client == nil {
			c.Next()
			return
		}

		bucket := time.Now().UnixNano() / int64(window)
		key := fmt.Sprintf("ratelimit:%s:%s:%d", scope, c.ClientIP(), bucket)

		ctx, cancel := context.WithTimeout(c.Request.Context(), 250*time.Millisecond)
		value, err := client.Incr(ctx, key).Result()
		if err == nil && value == 1 {
			_ = client.Expire(ctx, key, window).Err()
		}
		cancel()

		if err != nil {
			logrus.WithError(err).WithField("scope", scope).Warn("rate limiter unavailable; allowing request")
			c.Next()
			return
		}

		remaining := int64(limit) - value
		if remaining < 0 {
			remaining = 0
		}
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))

		if value > int64(limit) {
			c.Header("Retry-After", strconv.Itoa(int(window.Seconds())))
			c.Error(appErrors.NewAppError("Too many requests. Please try again shortly.", http.StatusTooManyRequests, nil))
			c.Abort()
			return
		}

		c.Next()
	}
}
