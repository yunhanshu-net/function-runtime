package natsx

import (
	"context"
	"fmt"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"go.uber.org/zap"
	"time"
)

// NATS相关配置常量
const (
	NatsConnectionTimeout = 5 * time.Second
	NatsServerPort        = 4222
	NatsMaxReconnects     = 10
	NatsReconnectWait     = 1 * time.Second
)

// InitNatsWithRetry 初始化NATS服务器和客户端，支持重试
func InitNatsWithRetry(ctx context.Context, maxRetries int) (*nats.Conn, *server.Server, error) {
	var natsSrv *server.Server
	var natsCli *nats.Conn
	var err error

	for i := 0; i < maxRetries; i++ {
		// 启动NATS服务器
		natsSrv, err = server.NewServer(&server.Options{
			Port: NatsServerPort,
		})
		if err != nil {
			logger.Warn(ctx, "NATS初始化失败",
				zap.Int("attempt", i+1),
				zap.Int("max_attempts", maxRetries),
				zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		// 启动服务器
		go natsSrv.Start()
		if !natsSrv.ReadyForConnections(10 * time.Second) {
			logger.Warn(ctx, "NATS服务器启动超时，等待重试...")
			natsSrv.Shutdown()
			time.Sleep(time.Second)
			continue
		}

		// 连接NATS服务器
		natsCli, err = nats.Connect(fmt.Sprintf("nats://localhost:%d", NatsServerPort),
			nats.ErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
				logger.Error(ctx, "NATS错误", err)
			}),
			nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
				logger.Errorf(ctx, "NATS err:,%v", err)
			}),
			nats.ReconnectHandler(func(conn *nats.Conn) {
				logger.Info(ctx, "NATS已重新连接")
			}),
		)
		if err != nil {
			logger.Warn(ctx, "NATS连接失败",
				zap.Int("attempt", i+1),
				zap.Int("max_attempts", maxRetries),
				zap.Error(err))
			natsSrv.Shutdown()
			time.Sleep(time.Second)
			continue
		}

		logger.Info(ctx, "NATS服务器已启动", zap.Int("port", NatsServerPort))
		return natsCli, natsSrv, nil
	}

	return nil, nil, fmt.Errorf("NATS初始化失败，已尝试%d次", maxRetries)
}
