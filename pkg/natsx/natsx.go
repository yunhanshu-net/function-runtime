package natsx

import (
	"fmt"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
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

	logrus.Infof("NATS服务器已启动，监听端口: %d", NatsServerPort)

	// 连接到服务器
	nc, err := nats.Connect(
		ns.ClientURL(),
		nats.MaxReconnects(NatsMaxReconnects),
		nats.ReconnectWait(NatsReconnectWait),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logrus.Warnf("NATS连接断开: %v", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logrus.Info("NATS已重新连接")
		}),
		nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
			logrus.Errorf("NATS错误: %v", err)
		}),
	)

	if err != nil {
		ns.Shutdown() // 关闭服务器
		return nil, nil, fmt.Errorf("连接到NATS服务器失败: %w", err)
	}

	logrus.Info("已成功连接到NATS服务器")
	return nc, ns, nil
}

// InitNatsWithRetry 初始化NATS服务，支持重试
func InitNatsWithRetry(maxRetries int) (*nats.Conn, *server.Server, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		natsCli, natsSrv, err := InitNats()
		if err == nil {
			// 连接成功
			return natsCli, natsSrv, nil
		}

		lastErr = err
		logrus.Warnf("NATS初始化失败(尝试 %d/%d): %v, 等待重试...", i+1, maxRetries, err)
		time.Sleep(time.Second * 2)
	}

	return nil, nil, fmt.Errorf("在%d次尝试后NATS初始化失败: %w", maxRetries, lastErr)
}
