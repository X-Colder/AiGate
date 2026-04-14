// AiGate - AI 网关服务
// 统一多个 AI 服务提供者（OpenAI、Anthropic、DeepSeek、豆包、通义千问、Kimi）的接口，
// 通过配置文件动态启用/禁用提供者，对外提供统一的 HTTP API。
package main

import (
	"log"

	"github.com/aigate/config"
	"github.com/aigate/pkg/logger"
	"github.com/aigate/router"
)

func main() {
	// 1. 加载配置（config.yaml 或环境变量指定的路径）
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化日志系统
	logger.Init(cfg.LogLevel)

	// 3. 初始化路由（含 Provider、Service、Handler 的依赖注入）
	r := router.Setup(cfg)

	// 4. 启动 HTTP 服务
	addr := ":" + cfg.Server.Port
	logger.Infof("AiGate server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
