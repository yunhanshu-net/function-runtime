package kernel

import (
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/pkg/filecache"
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
	executor, err := NewDefaultExecutor()
	if err != nil {
		return nil, err
	}
	return &Runcher{executor: executor}, nil
}
