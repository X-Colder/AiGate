// AiGate - AI 网关服务
// 支持多租户、网关管理（熔断/限流/降级）、监控统计、Vue3 前端管理界面
package main

import (
	"log"

	"github.com/aigate/config"
	"github.com/aigate/pkg/logger"
	"github.com/aigate/router"
	"github.com/aigate/store"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化日志
	logger.Init(cfg.LogLevel)

	// 3. 初始化数据库
	dbPath := cfg.Database.Path
	if dbPath == "" {
		dbPath = "aigate.db"
	}
	if err := store.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 4. 初始化默认租户（首次启动时创建）
	if err := store.InitDefaultTenant(); err != nil {
		logger.Warnf("Init default tenant: %v", err)
	}

	// 5. 初始化路由
	r := router.Setup(cfg, store.GetDB())

	// 6. 启动 HTTP 服务
	addr := ":" + cfg.Server.Port
	logger.Infof("AiGate server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
