package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aigate/pkg/response"
)

// RequireTenantAccess 要求租户管理权限
func RequireTenantAccess() gin.HandlerFunc {
	return requirePermission("tenant_access")
}

// RequireGatewayAccess 要求网关管理权限
func RequireGatewayAccess() gin.HandlerFunc {
	return requirePermission("gateway_access")
}

// RequireMonitorAccess 要求监控面板权限
func RequireMonitorAccess() gin.HandlerFunc {
	return requirePermission("monitor_access")
}

// requirePermission 通用权限校验中间件
// 从 gin.Context 中读取 Auth 中间件注入的权限字段
func requirePermission(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get(key)
		if !exists || val != true {
			response.Error(c, http.StatusForbidden, "no permission: "+key)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAnyPermission 要求至少拥有一种指定权限
func RequireAnyPermission(keys ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, key := range keys {
			val, exists := c.Get(key)
			if exists && val == true {
				c.Next()
				return
			}
		}
		response.Error(c, http.StatusForbidden, "insufficient permissions")
		c.Abort()
	}
}
