package kernel

import (
	"context"
	"fmt"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/kernel/coder"
	"github.com/yunhanshu-net/runcher/kernel/scheduler"
	"sync"
	"time"
)

// Runcher 是核心运行器，管理调度器和编码器
type Runcher struct {
	Scheduler  *scheduler.Scheduler
	Coder      *coder.Coder
	natsServer *server.Server
	natsConn   *nats.Conn
	down       chan struct{}
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
}

// MustNewRuncher 创建一个新的Runcher实例
// 如果初始化过程中出现任何错误，将会panic
func MustNewRuncher() *Runcher {
	natsCli, natsSrv, err := InitNatsWithRetry(3)
	if err != nil {
		panic(fmt.Sprintf("初始化NATS失败: %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Runcher{
		down:       make(chan struct{}, 1),
		natsServer: natsSrv,
		natsConn:   natsCli,
		Scheduler:  scheduler.NewScheduler(natsCli),
		Coder:      coder.NewDefaultCoder(natsCli),
		ctx:        ctx,
		cancel:     cancel,
	}
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

// Run 启动Runcher
func (a *Runcher) Run() error {
	logrus.Info("启动Runcher...")

	// 启动调度器
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		err := a.Scheduler.Run()
		if err != nil {
			logrus.Errorf("调度器运行错误: %v", err)
			a.cancel() // 出错时取消所有组件
		}
	}()

	// 启动编码器
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		err := a.Coder.Run()
		if err != nil {
			logrus.Errorf("编码器运行错误: %v", err)
			a.cancel() // 出错时取消所有组件
		}
	}()

	// 监控上下文取消
	go func() {
		<-a.ctx.Done()
		logrus.Info("接收到取消信号，准备关闭Runcher...")
		close(a.down)
	}()

	logrus.Info("Runcher启动成功")
	return nil
}

// Close 关闭Runcher及其所有组件
func (a *Runcher) Close() error {
	logrus.Info("开始关闭Runcher...")

	// 发送取消信号
	a.cancel()

	// 设置超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 等待关闭或超时
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 正常关闭
	case <-ctx.Done():
		logrus.Warn("Runcher关闭超时")
	}

	// 关闭组件
	var errs []error

	if err := a.Coder.Close(); err != nil {
		errs = append(errs, fmt.Errorf("关闭编码器错误: %w", err))
	}

	if err := a.Scheduler.Close(); err != nil {
		errs = append(errs, fmt.Errorf("关闭调度器错误: %w", err))
	}

	// 关闭NATS连接
	if a.natsConn != nil {
		a.natsConn.Close()
	}

	// 停止NATS服务器
	if a.natsServer != nil {
		a.natsServer.Shutdown()
	}

	logrus.Info("Runcher已完全关闭")

	if len(errs) > 0 {
		return fmt.Errorf("关闭过程中发生%d个错误: %v", len(errs), errs)
	}

	return nil
}
