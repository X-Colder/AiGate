package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/aigate/model"
	"github.com/aigate/pkg/auth"
	"github.com/aigate/pkg/response"
)

// Auth JWT 认证中间件，从 Authorization 头提取 Bearer Token 并验证。
// 同时根据用户关联的 Role 将模块权限注入 gin.Context，供 Permission 中间件使用。
func Auth(db ...*gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "invalid authorization format")
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		// 将用户基础信息注入 gin.Context
		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		// 查询用户角色权限并注入 context（供 Permission 中间件使用）
		if len(db) > 0 && db[0] != nil {
			var user model.User
			if db[0].Where("id = ?", claims.UserID).First(&user).Error == nil && user.RoleID != "" {
				var role model.Role
				if db[0].Where("id = ?", user.RoleID).First(&role).Error == nil {
					c.Set("tenant_access", role.TenantAccess)
					c.Set("gateway_access", role.GatewayAccess)
					c.Set("monitor_access", role.MonitorAccess)
				}
			} else if claims.Role == "admin" {
				// 兼容旧用户无 RoleID
				c.Set("tenant_access", true)
				c.Set("gateway_access", true)
				c.Set("monitor_access", true)
			}
		}

		c.Next()
	}
}
