package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openclaw/openclaw-saas-platform/internal/api"
	"github.com/openclaw/openclaw-saas-platform/pkg/config"
	"github.com/openclaw/openclaw-saas-platform/pkg/logger"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log, err := logger.NewLogger(cfg.Server.Mode)
	if err != nil {
		fmt.Printf("初始化日志失败: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("OpenClaw SaaS 平台启动中...")
	log.Info(fmt.Sprintf("运行模式: %s", cfg.Server.Mode))
	log.Info(fmt.Sprintf("监听端口: %d", cfg.Server.Port))

	// 初始化路由
	router, err := api.NewRouter(cfg, log)
	if err != nil {
		log.Fatal(fmt.Sprintf("初始化路由失败: %v", err))
	}

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// 优雅关闭：在后台启动服务器
	go func() {
		log.Info(fmt.Sprintf("HTTP 服务器启动，监听 :%d", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(fmt.Sprintf("HTTP 服务器启动失败: %v", err))
		}
	}()

	// 等待中断信号（SIGINT 或 SIGTERM）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("收到关闭信号，正在优雅关闭服务器...")

	// 设置关闭超时时间为 30 秒
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(fmt.Sprintf("服务器强制关闭: %v", err))
	}

	log.Info("服务器已成功关闭")
}
