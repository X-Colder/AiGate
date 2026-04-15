// Package router 初始化 Gin 路由引擎，注册中间件和 API 路由。
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/aigate/config"
	"github.com/aigate/handler"
	"github.com/aigate/middleware"
	"github.com/aigate/provider"
	"github.com/aigate/service"
)

// Setup 根据配置初始化并返回 Gin 路由引擎。
func Setup(cfg *config.Config, db *gorm.DB) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	r := gin.New()

	// 全局中间件链：异常恢复 → 请求日志 → 跨域处理
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())

	// ===== 依赖初始化 =====
	// Provider (原有聊天转发功能)
	registry := provider.InitProviders(cfg)
	chatService := service.NewChatService(registry)
	chatHandler := handler.NewChatHandler(chatService)

	// 认证
	authService := service.NewAuthService(db)
	authHandler := handler.NewAuthHandler(authService)

	// 网关管理
	gatewayService := service.NewGatewayService(db)
	gatewayHandler := handler.NewGatewayHandler(gatewayService)

	// 监控指标
	metricService := service.NewMetricService(db)
	metricHandler := handler.NewMetricHandler(metricService)

	// 租户管理（仅管理员）
	tenantService := service.NewTenantService(db)
	tenantHandler := handler.NewTenantHandler(tenantService)

	// ===== 前端静态文件服务 =====
	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/", "./frontend/dist/index.html")
	r.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")
	// Vue Router history 模式兜底：非 API 路径返回 index.html
	r.NoRoute(func(c *gin.Context) {
		// API 路径返回 404 JSON
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "not found"})
			return
		}
		c.File("./frontend/dist/index.html")
	})

	// ===== 健康检查 =====
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "AiGate"})
	})

	// ===== API v1 路由 =====
	api := r.Group("/api/v1")
	{
		// 公开接口：认证
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
		}

		// 需要认证的接口
		protected := api.Group("")
		protected.Use(middleware.Auth())
		{
			// 聊天
			protected.POST("/chat", chatHandler.Chat)
			protected.GET("/providers", chatHandler.ListProviders)

			// 网关管理 CRUD
			protected.GET("/gateways", gatewayHandler.List)
			protected.POST("/gateways", gatewayHandler.Create)
			protected.GET("/gateways/:id", gatewayHandler.GetByID)
			protected.PUT("/gateways/:id", gatewayHandler.Update)
			protected.DELETE("/gateways/:id", gatewayHandler.Delete)
			protected.PUT("/gateways/:id/policy", gatewayHandler.UpdatePolicy)

			// 监控指标
			protected.GET("/metrics/summary", metricHandler.GetSummary)
			protected.GET("/metrics/trend", metricHandler.GetTrend)
		}

		// 管理员专用接口（需认证 + admin 角色）
		admin := api.Group("/admin")
		admin.Use(middleware.Auth(), middleware.AdminOnly())
		{
			admin.GET("/tenants", tenantHandler.List)
			admin.POST("/tenants", tenantHandler.Create)
			admin.GET("/tenants/:id", tenantHandler.GetByID)
			admin.PUT("/tenants/:id", tenantHandler.Update)
			admin.DELETE("/tenants/:id", tenantHandler.Delete)
			admin.GET("/tenants/:id/usage", tenantHandler.GetUsage)
		}
	}

	return r
}
