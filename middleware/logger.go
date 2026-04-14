package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/aigate/pkg/logger"
)

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Infof("[%d] %s %s %v", statusCode, c.Request.Method, path, latency)
	}
}
