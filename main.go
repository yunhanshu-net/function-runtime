package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/function-runtime/cmd"
	"github.com/yunhanshu-net/function-runtime/pkg/config"
	"github.com/yunhanshu-net/function-runtime/router"
	"github.com/yunhanshu-net/pkg/logger"
)

func main() {
	// 加载配置
	if err := config.Init(); err != nil {
		log.Fatalf("初始化配置失败: %v", err)
	}

	// 初始化日志
	cfg := config.Get()
	logCfg := logger.Config{
		Level:      cfg.LogConfig.Level,
		Filename:   cfg.LogConfig.Filename,
		MaxSize:    cfg.LogConfig.MaxSize,
		MaxBackups: cfg.LogConfig.MaxBackups,
		MaxAge:     cfg.LogConfig.MaxAge,
		Compress:   cfg.LogConfig.Compress,
		IsDev:      cfg.ServerConfig.Mode == "debug",
	}
	if err := logger.Init(logCfg); err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	ctx := context.Background()

	// 初始化应用组件
	cmd.Init()
	defer func() {
		if cmd.Runcher != nil {
			cmd.Runcher.Close()
		}
	}()

	app := gin.New()
	router.InitRouter(app)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    "0.0.0.0:9999",
		Handler: app,
	}

	// 启动HTTP服务
	go func() {
		logger.Infof(ctx, "HTTP服务启动成功，监听端口: 9999")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(ctx, "HTTP服务启动失败", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info(ctx, "正在关闭服务...")

	// 5秒超时关闭服务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error(ctx, "服务关闭时出错", err)
	}

	logger.Info(ctx, "服务已安全关闭")
}
