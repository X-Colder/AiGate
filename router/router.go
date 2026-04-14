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

// Setup 初始化路由
func Setup(cfg *config.Config) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	r := gin.New()

	// 全局中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())

	// 初始化 providers
	registry := provider.InitProviders(cfg)

	// 初始化 services
	chatService := service.NewChatService(registry)

	// 初始化 handlers
	chatHandler := handler.NewChatHandler(chatService)

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "AiGate",
		})
	})

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 聊天相关
		api.POST("/chat", chatHandler.Chat)
		// 提供者列表
		api.GET("/providers", chatHandler.ListProviders)
	}

	return r
}
