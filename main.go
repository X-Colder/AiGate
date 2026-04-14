package main

import (
	"log"

	"github.com/aigate/config"
	"github.com/aigate/pkg/logger"
	"github.com/aigate/router"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger.Init(cfg.LogLevel)

	// 初始化路由
	r := router.Setup(cfg)

	// 启动服务
	addr := ":" + cfg.Server.Port
	logger.Infof("AiGate server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
