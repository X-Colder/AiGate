// Package router 初始化 Gin 路由引擎，注册中间件和 API 路由。
// 负责将 HTTP 请求分发到对应的 Handler 处理。
package router

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/aigate/config"
	"github.com/aigate/handler"
	"github.com/aigate/middleware"
	"github.com/aigate/provider"
	"github.com/aigate/service"
)

// Setup 根据配置初始化并返回 Gin 路由引擎。
// 执行流程：设置 Gin 模式 → 注册全局中间件 → 初始化 Provider/Service/Handler → 注册路由。
func Setup(cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	r := gin.New()

	// 全局中间件链：异常恢复 → 请求日志 → 跨域处理
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())

	// 依赖初始化链：Config → Provider Registry → Service → Handler
	registry := provider.InitProviders(cfg)
	chatService := service.NewChatService(registry)
	chatHandler := handler.NewChatHandler(chatService)

	// 健康检查端点，用于负载均衡器和监控探测
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "AiGate",
		})
	})

	// API v1 路由组
	api := r.Group("/api/v1")
	{
		api.POST("/chat", chatHandler.Chat)              // 聊天对话（支持普通/流式）
		api.GET("/providers", chatHandler.ListProviders) // 获取已启用的提供者列表
	}

	return r
}
