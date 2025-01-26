package kernel

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/pkg/filecache"
	"github.com/yunhanshu-net/runcher/pkg/natsx"
	"github.com/yunhanshu-net/runcher/pkg/store"
)

type Runcher struct {
	upstream      *nats.Conn //上游输入
	upstreamSub   *nats.Subscription
	downstream    *nats.Conn //下游数据交互
	downstreamSub *nats.Subscription
	fileInCache   filecache.FileCache //输入数据缓存
	fileOutCache  filecache.FileCache //输出数据缓存

	executor *Executor //执行软件
}

func NewDefaultRuncher() (*Runcher, error) {

	r := &Runcher{}
	upstreamSubKey := fmt.Sprintf("runner.run.*.*.*")
	upstreamSub, err := natsx.Nats.Subscribe(upstreamSubKey, r.upstreamFunc)
	if err != nil {
		return nil, err
	}
	downstreamSubKey := fmt.Sprintf("runner.request.*.*.*")
	downstreamSub, err := natsx.Nats.Subscribe(downstreamSubKey, r.downstreamFunc)
	if err != nil {
		return nil, err
	}
	r = &Runcher{
		upstreamSub:   upstreamSub,
		downstreamSub: downstreamSub,
		upstream:      natsx.Nats,
		downstream:    natsx.Nats,
		fileInCache:   filecache.NewLocalFileCache(),
		fileOutCache:  filecache.NewLocalFileCache(),
		executor:      NewExecutor(store.NewDefaultQiNiu()),
	}
	err = r.executor.RunnerListen()
	if err != nil {
		return nil, err
	}
	return r, nil
}
