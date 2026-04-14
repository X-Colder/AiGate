package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aigate/pkg/logger"
	"github.com/aigate/pkg/response"
)

// Recovery 恢复中间件，捕获 panic
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Errorf("Panic recovered: %v", err)
				response.Error(c, http.StatusInternalServerError, "Internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
