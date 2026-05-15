// Package router 初始化 Gin 路由引擎，注册中间件和 API 路由。
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/aigate/config"
	"github.com/aigate/handler"
	"github.com/aigate/middleware"
	"github.com/aigate/pkg/resilience"
	"github.com/aigate/provider"
	"github.com/aigate/service"
	"github.com/aigate/store"
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
	// 熔断器
	breaker := resilience.NewCircuitBreaker(store.GetRedis())

	// Provider (原有聊天转发功能)
	registry := provider.InitProviders(cfg)
	chatService := service.NewChatService(registry, breaker, db)
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

	// 角色管理（仅管理员）
	roleService := service.NewRoleService(db)
	roleHandler := handler.NewRoleHandler(roleService)

	// 用户管理（仅管理员）
	userService := service.NewUserService(db)
	userHandler := handler.NewUserHandler(userService)

	// C端服务
	modelService := service.NewModelService(db)
	billingService := service.NewBillingService(db)
	apikeyService := service.NewAPIKeyService(db)
	usageService := service.NewUsageService(db)

	modelHandler := handler.NewModelHandler(modelService, billingService)
	apikeyHandler := handler.NewAPIKeyHandler(apikeyService)
	developerHandler := handler.NewDeveloperHandler(modelService, billingService, usageService)
	inferenceHandler := handler.NewInferenceHandler(modelService, billingService, usageService, apikeyService, registry, db)

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
		health := gin.H{"status": "ok", "service": "AiGate"}
		status := http.StatusOK

		// 检查数据库
		if sqlDB, err := db.DB(); err != nil || sqlDB.Ping() != nil {
			health["status"] = "degraded"
			health["mysql"] = "down"
			status = http.StatusServiceUnavailable
		} else {
			health["mysql"] = "ok"
		}

		// 检查 Redis
		rdb := store.GetRedis()
		if rdb != nil {
			if rdb.Ping(c.Request.Context()).Err() != nil {
				health["status"] = "degraded"
				health["redis"] = "down"
				status = http.StatusServiceUnavailable
			} else {
				health["redis"] = "ok"
			}
		}

		c.JSON(status, health)
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

		// 需要认证的接口（Auth 中间件注入权限信息）
		protected := api.Group("")
		protected.Use(middleware.Auth(db))
		{
			// 限流中间件
			rateLimiter := middleware.NewRateLimiter(store.GetRedis(), db)

			// 聊天（带限流）
			protected.POST("/chat", middleware.RateLimit(rateLimiter), chatHandler.Chat)
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

		// 管理员专用接口（需认证 + 租户管理权限）
		admin := api.Group("/admin")
		admin.Use(middleware.Auth(db), middleware.RequireTenantAccess())
		{
			admin.GET("/tenants", tenantHandler.List)
			admin.POST("/tenants", tenantHandler.Create)
			admin.GET("/tenants/:id", tenantHandler.GetByID)
			admin.PUT("/tenants/:id", tenantHandler.Update)
			admin.DELETE("/tenants/:id", tenantHandler.Delete)
			admin.GET("/tenants/:id/usage", tenantHandler.GetUsage)

			admin.GET("/roles", roleHandler.List)
			admin.POST("/roles", roleHandler.Create)
			admin.GET("/roles/:id", roleHandler.GetByID)
			admin.PUT("/roles/:id", roleHandler.Update)
			admin.DELETE("/roles/:id", roleHandler.Delete)

			admin.GET("/users", userHandler.List)
			admin.POST("/users", userHandler.Create)
			admin.GET("/users/:id", userHandler.GetByID)
			admin.PUT("/users/:id", userHandler.Update)
			admin.DELETE("/users/:id", userHandler.Delete)

			// 模型管理
			admin.GET("/models", modelHandler.List)
			admin.POST("/models", modelHandler.Create)
			admin.GET("/models/:id", modelHandler.GetByID)
			admin.PUT("/models/:id", modelHandler.Update)
			admin.DELETE("/models/:id", modelHandler.Delete)
			admin.PUT("/models/:id/doc", modelHandler.UpdateDoc)

			// 计费管理
			admin.GET("/billing/users", modelHandler.ListUserBalances)
			admin.POST("/billing/recharge", modelHandler.Recharge)
			admin.GET("/billing/models", modelHandler.GetFinanceSummary)
			admin.POST("/billing/models/recharge", modelHandler.RechargeModel)
			admin.GET("/billing/models/recharge-history", modelHandler.GetRechargeHistory)
		}

		// C端开发者接口（需认证）
		developer := api.Group("/developer")
		developer.Use(middleware.Auth(db))
		{
			developer.POST("/apikeys", apikeyHandler.Create)
			developer.GET("/apikeys", apikeyHandler.List)
			developer.PUT("/apikeys/:id", apikeyHandler.Update)
			developer.DELETE("/apikeys/:id", apikeyHandler.Delete)

			developer.GET("/models", developerHandler.ListModels)
			developer.GET("/models/:id/doc", developerHandler.GetModelDoc)
			developer.GET("/balance", developerHandler.GetBalance)
			developer.GET("/transactions", developerHandler.GetTransactions)
			developer.GET("/usage/summary", developerHandler.GetUsageSummary)
			developer.GET("/usage/trend", developerHandler.GetUsageTrend)
			developer.GET("/usage/records", developerHandler.GetUsageRecords)
		}
	}

	// ===== OpenAI 兼容推理接口 (API Key 认证) =====
	v1Open := r.Group("/v1")
	v1Open.Use(middleware.APIKeyAuth(db))
	{
		v1Open.POST("/chat/completions", inferenceHandler.ChatCompletion)
		v1Open.GET("/models", inferenceHandler.ListModels)
	}

	return r
}
