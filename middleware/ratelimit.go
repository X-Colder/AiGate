package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/aigate/model"
	"github.com/aigate/pkg/response"
)

type RateLimiter struct {
	rdb *redis.Client
	db  *gorm.DB
}

func NewRateLimiter(rdb *redis.Client, db *gorm.DB) *RateLimiter {
	return &RateLimiter{rdb: rdb, db: db}
}

func (rl *RateLimiter) Allow(ctx context.Context, key string, qps int, burst int) (bool, error) {
	if rl.rdb == nil {
		return true, nil
	}

	now := time.Now().UnixNano()
	windowSize := int64(time.Second)
	windowStart := now - windowSize

	pipe := rl.rdb.Pipeline()
	// 移除窗口外的请求记录
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))
	// 统计窗口内请求数
	countCmd := pipe.ZCard(ctx, key)
	// 添加当前请求
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now})
	// 设置键过期
	pipe.Expire(ctx, key, 2*time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return true, err
	}

	count := countCmd.Val()
	limit := int64(qps)
	if burst > qps {
		limit = int64(burst)
	}

	return count < limit, nil
}

func RateLimit(rl *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rl == nil || rl.rdb == nil {
			c.Next()
			return
		}

		tenantID, _ := c.Get("tenant_id")
		if tenantID == nil {
			c.Next()
			return
		}

		// 查找该租户下启用了限流的网关
		var policies []struct {
			GatewayID    string
			RateLimitQPS   int
			RateLimitBurst int
		}
		rl.db.Model(&model.GatewayPolicy{}).
			Select("gateway_policies.gateway_id, gateway_policies.rate_limit_qps, gateway_policies.rate_limit_burst").
			Joins("JOIN gateways ON gateways.id = gateway_policies.gateway_id").
			Where("gateways.tenant_id = ? AND gateway_policies.rate_limit_enabled = ?", tenantID, true).
			Scan(&policies)

		if len(policies) == 0 {
			c.Next()
			return
		}

		// 使用该租户最严格的限流策略
		qps := policies[0].RateLimitQPS
		burst := policies[0].RateLimitBurst
		for _, p := range policies[1:] {
			if p.RateLimitQPS < qps {
				qps = p.RateLimitQPS
				burst = p.RateLimitBurst
			}
		}

		key := fmt.Sprintf("ratelimit:%s", tenantID)
		allowed, err := rl.Allow(c.Request.Context(), key, qps, burst)
		if err != nil {
			c.Next()
			return
		}

		if !allowed {
			response.Error(c, http.StatusTooManyRequests, "rate limit exceeded")
			c.Abort()
			return
		}

		c.Next()
	}
}
