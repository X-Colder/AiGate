package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/aigate/model"
	"github.com/aigate/pkg/auth"
	"github.com/aigate/pkg/response"
	"github.com/aigate/store"
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

		c.Set("user_id", claims.UserID)
		c.Set("tenant_id", claims.TenantID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		if len(db) > 0 && db[0] != nil {
			loadPermissions(c, db[0], claims)
		}

		c.Next()
	}
}

func loadPermissions(c *gin.Context, db *gorm.DB, claims *auth.Claims) {
	var user model.User
	if db.Where("id = ?", claims.UserID).First(&user).Error != nil || user.RoleID == "" {
		if claims.Role == "admin" {
			c.Set("tenant_access", true)
			c.Set("gateway_access", true)
			c.Set("monitor_access", true)
		}
		return
	}

	// Redis 缓存优先
	cached, err := store.GetCachedRolePermissions(c.Request.Context(), user.RoleID)
	if err == nil && cached != nil {
		c.Set("tenant_access", cached.TenantAccess)
		c.Set("gateway_access", cached.GatewayAccess)
		c.Set("monitor_access", cached.MonitorAccess)
		return
	}

	// 缓存未命中，查 DB
	var role model.Role
	if db.Where("id = ?", user.RoleID).First(&role).Error == nil {
		c.Set("tenant_access", role.TenantAccess)
		c.Set("gateway_access", role.GatewayAccess)
		c.Set("monitor_access", role.MonitorAccess)

		// 写回缓存
		_ = store.CacheRolePermissions(c.Request.Context(), user.RoleID, &store.RolePermissions{
			TenantAccess:  role.TenantAccess,
			GatewayAccess: role.GatewayAccess,
			MonitorAccess: role.MonitorAccess,
		})
	}
}

