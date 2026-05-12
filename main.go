package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aigate/config"
	"github.com/aigate/pkg/auth"
	"github.com/aigate/pkg/logger"
	"github.com/aigate/router"
	"github.com/aigate/service"
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

	// 3. 注入 JWT Secret
	if cfg.JWTSecret != "" {
		auth.SetSecret(cfg.JWTSecret)
	}

	// 4. 初始化数据库
	if err := store.InitDB(&cfg.Database); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	// 5. 初始化 Redis（可选，不可用时降级为无缓存模式）
	if cfg.Redis.Addr != "" {
		if err := store.InitRedis(&cfg.Redis); err != nil {
			logger.Warnf("Redis not available, running without cache: %v", err)
		}
	}

	// 6. 初始化默认租户（首次启动时创建）
	if err := store.InitDefaultTenant(); err != nil {
		logger.Warnf("Init default tenant: %v", err)
	}

	// 7. 启动 Metrics 缓冲写入
	metricBuffer := service.NewMetricBuffer(store.GetDB(), 100, 5*time.Second)
	metricBuffer.Start()

	// 8. 初始化路由
	r := router.Setup(cfg, store.GetDB())

	// 9. 启动 HTTP 服务（优雅停机模式）
	addr := ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Infof("AiGate server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 10. 等待关闭信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Infof("Shutting down server...")

	// 30 秒内完成在途请求
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Errorf("Server shutdown error: %v", err)
	}

	// 刷写缓冲 Metrics
	metricBuffer.Stop()

	// 关闭 Redis
	store.CloseRedis()

	// 关闭数据库
	if sqlDB, err := store.GetDB().DB(); err == nil {
		sqlDB.Close()
	}

	logger.Infof("Server exited")
}
