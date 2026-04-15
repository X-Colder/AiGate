package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aigate/pkg/response"
)

// AdminOnly 管理员权限中间件，仅允许 role=admin 的用户访问
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("role")
		if role != "admin" {
			response.Error(c, http.StatusForbidden, "admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
