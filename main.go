package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/runcher/cmd"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"github.com/yunhanshu-net/runcher/router"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	// 初始化日志系统，已在zap.go的init()中完成，这里不需要显式调用
	// logger.Setup() 是logrus.go中的方法，我们不再需要

	// 初始化应用组件
	cmd.Init()

	defer cmd.Runcher.Close()

	app := gin.New()
	router.InitRouter(app)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    "0.0.0.0:9999",
		Handler: app,
	}

	// 启动HTTP服务
	go func() {
		logger.Info("HTTP服务启动成功，监听端口: 9999")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("HTTP服务启动失败", zap.String("error", err.Error()))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("正在关闭服务...")

	// 5秒超时关闭服务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("服务关闭时出错", zap.String("error", err.Error()))
	}

	logger.Info("服务已安全关闭")
}
