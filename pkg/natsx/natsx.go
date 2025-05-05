package natsx

import (
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

// InitNats 初始化NATS服务器和客户端连接
// 返回NATS客户端连接、服务器实例和可能的错误
func InitNats() (*nats.Conn, *server.Server, error) {
	// 配置NATS服务器选项
	opts := &server.Options{
		Port:           NatsServerPort,
		NoLog:          false,
		NoSigs:         true,
		MaxControlLine: 4096,
	}

	// 初始化新服务器
	ns, err := server.NewServer(opts)
	if err != nil {
		return nil, nil, fmt.Errorf("创建NATS服务器失败: %w", err)
	}

	// 异步启动服务器
	go ns.Start()

	// 等待服务器准备就绪
	if !ns.ReadyForConnections(NatsConnectionTimeout) {
		ns.Shutdown() // 关闭服务器
		return nil, nil, fmt.Errorf("NATS服务器未能在%s内准备就绪", NatsConnectionTimeout)
	}

	logger.Info("NATS服务器已启动", zap.Int("port", NatsServerPort))

	// 连接到服务器
	nc, err := nats.Connect(
		ns.ClientURL(),
		nats.MaxReconnects(NatsMaxReconnects),
		nats.ReconnectWait(NatsReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Warn("NATS连接断开", zap.Error(err))
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Info("NATS已重新连接")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			logger.Error("NATS错误", zap.Error(err))
		}),
	)

	if err != nil {
		ns.Shutdown() // 关闭服务器
		return nil, nil, fmt.Errorf("连接到NATS服务器失败: %w", err)
	}

	logger.Info("已成功连接到NATS服务器")
	return nc, ns, nil
}

// InitNatsWithRetry 初始化NATS服务器和客户端，支持重试
func InitNatsWithRetry(maxRetries int) (*nats.Conn, *server.Server, error) {
	var natsSrv *server.Server
	var natsCli *nats.Conn
	var err error

	for i := 0; i < maxRetries; i++ {
		// 启动NATS服务器
		natsSrv, err = server.NewServer(&server.Options{
			Port: NatsServerPort,
		})
		if err != nil {
			logger.Warn("NATS初始化失败",
				zap.Int("attempt", i+1),
				zap.Int("max_attempts", maxRetries),
				zap.Error(err))
			time.Sleep(time.Second)
			continue
		}

		// 启动服务器
		go natsSrv.Start()
		if !natsSrv.ReadyForConnections(10 * time.Second) {
			logger.Warn("NATS服务器启动超时，等待重试...")
			natsSrv.Shutdown()
			time.Sleep(time.Second)
			continue
		}

		// 连接NATS服务器
		natsCli, err = nats.Connect(fmt.Sprintf("nats://localhost:%d", NatsServerPort),
			nats.ErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
				logger.Error("NATS错误", zap.Error(err))
			}),
			nats.DisconnectHandler(func(conn *nats.Conn) {
				logger.Warn("NATS连接断开")
			}),
			nats.ReconnectHandler(func(conn *nats.Conn) {
				logger.Info("NATS已重新连接")
			}),
		)
		if err != nil {
			logger.Warn("NATS连接失败",
				zap.Int("attempt", i+1),
				zap.Int("max_attempts", maxRetries),
				zap.Error(err))
			natsSrv.Shutdown()
			time.Sleep(time.Second)
			continue
		}

		logger.Info("NATS服务器已启动", zap.Int("port", NatsServerPort))
		return natsCli, natsSrv, nil
	}

	return nil, nil, fmt.Errorf("NATS初始化失败，已尝试%d次", maxRetries)
}
