package main

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/cmd"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"github.com/yunhanshu-net/runcher/router"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//conns.Init()
	logger.Setup()
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
		logrus.Info("HTTP服务启动成功，监听端口: 9999")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.Fatalf("HTTP服务启动失败: %s", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("正在关闭服务...")

	// 5秒超时关闭服务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Errorf("服务关闭时出错: %s", err)
	}

	logrus.Info("服务已安全关闭")
}
