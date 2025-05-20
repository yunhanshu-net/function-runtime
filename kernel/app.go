package kernel

import (
	"context"
	"fmt"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/pkg/x/natsx"
	"github.com/yunhanshu-net/runcher/kernel/scheduler"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"sync"
	"time"
)

// Runcher 是核心运行器，管理调度器和编码器
type Runcher struct {
	Scheduler *scheduler.Scheduler
	//Coder      *coder.Coder
	natsServer *server.Server
	natsConn   *nats.Conn
	down       chan struct{}
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc

	manageSub *nats.Subscription
}

func (a *Runcher) GetNatsConn() *nats.Conn {
	return a.natsConn
}

// MustNewRuncher 创建一个新的Runcher实例
// 如果初始化过程中出现任何错误，将会panic
func MustNewRuncher() *Runcher {

	ctx, cancel := context.WithCancel(context.Background())
	natsCli, natsSrv, err := natsx.InitNatsWithRetry(ctx, 3)
	if err != nil {
		panic(fmt.Sprintf("初始化NATS失败: %v", err))
	}

	return &Runcher{
		down:       make(chan struct{}, 1),
		natsServer: natsSrv,
		natsConn:   natsCli,
		Scheduler:  scheduler.NewScheduler(natsCli),
		//Coder:      coder.NewDefaultCoder(natsCli),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Run 启动Runcher
func (a *Runcher) Run() error {
	logger.Info(context.Background(), "启动Runcher...")

	// 启动调度器
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		err := a.Scheduler.Run()
		if err != nil {
			logger.Error(context.Background(), "调度器运行错误", err)
			a.cancel() // 出错时取消所有组件
		}
	}()

	//// 启动编码器
	//a.wg.Add(1)
	//go func() {
	//	defer a.wg.Done()
	//	err := a.Coder.Run()
	//	if err != nil {
	//		logger.Error("编码器运行错误", logger.Error(err))
	//		a.cancel() // 出错时取消所有组件
	//	}
	//}()

	// 监控上下文取消
	go func() {
		<-a.ctx.Done()
		logger.Info(context.Background(), "接收到取消信号，准备关闭Runcher...")
		close(a.down)
	}()

	logger.Info(context.Background(), "Runcher启动成功")
	return nil
}

// Close 关闭Runcher及其所有组件
func (a *Runcher) Close() error {
	logger.Info(context.Background(), "开始关闭Runcher...")

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
		logger.Warn(context.Background(), "Runcher关闭超时")
	}

	// 关闭组件
	var errs []error
	//
	//if err := a.Coder.Close(); err != nil {
	//	errs = append(errs, fmt.Errorf("关闭编码器错误: %w", err))
	//}

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

	logger.Info(context.Background(), "Runcher已完全关闭")

	if len(errs) > 0 {
		return fmt.Errorf("关闭过程中发生%d个错误: %v", len(errs), errs)
	}

	return nil
}
