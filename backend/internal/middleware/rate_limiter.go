package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/costguard/costguard/internal/constants"
	"github.com/costguard/costguard/internal/util"
)

// RateLimit is a Redis-backed fixed-window limiter (per IP, maxRequests/min).
func RateLimit(rdb *redis.Client, maxRequests int, logger *slog.Logger) gin.HandlerFunc {
	if maxRequests <= 0 {
		maxRequests = 60
	}
	return func(c *gin.Context) {
		ip := c.ClientIP()
		ctx := context.Background()
		key := fmt.Sprintf("costguard:ratelimit:%s:%d", ip, time.Now().Unix()/60)
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			logger.Warn("rate limiter error", "error", err)
			c.Next()
			return
		}
		if count == 1 {
			_ = rdb.Expire(ctx, key, 2*time.Minute).Err()
		}
		if count > int64(maxRequests) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, util.Response{
				Code:    constants.CodeTooManyRequests,
				Message: constants.MsgTooManyRequests,
			})
			return
		}
		c.Next()
	}
}
