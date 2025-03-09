package v2

import (
	"fmt"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/runtime"
	"sync"
	"time"
)

func (r *Runcher) request(ctx *request.Context) (*response.RunnerResponse, error) {
	subject := ctx.Request.GetSubject()
	ctx.Conn = r.conn
	r.lk.Lock()
	runners, ok := r.runners[subject]

	if !ok { //如果主题不存在，先初始化主题
		mp := make(map[int64]int)
		mp[time.Now().Unix()] = 1 //记录qps
		r.runners[subject] = &runtime.Runners{
			RWMutex:            &sync.RWMutex{},
			QPSRWMutex:         &sync.RWMutex{},
			StartingCountMutex: &sync.RWMutex{},
			QPS:                mp,
			CurrentPosition:    0,
			Running:            make([]*runtime.Runner, 0, 3),
		}
		r.lk.Unlock()
		//然后执行一次请求
		fmt.Printf("r.runRequest(ctx)")
		return r.runRequest(ctx)
	}
	r.lk.Unlock()
	runners.AddCurrentQPS(1)

	//runners.RWMutex.Lock()
	//defer runners.RWMutex.Unlock()

	runningCount := runners.GetRunningCount()
	qps := runners.GetCurrentQPSCount()
	if runningCount > 0 { //说明此时有运行中的实例,判断是否需要扩容
		addCount := qps - (2000 * runningCount)
		c := addCount / 2000
		if c > 0 { //需要扩容c台实例
			//runners.RWMutex.Lock()
			for i := 0; i < c; i++ {
				_, err := r.startNewRunner(ctx)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if runners.GetRunningCount() == 0 { //如果主题存在，且此时无长连接启动操作，需要启动实例，//////就需要判断有没有运行实例，没有运行实例，判断当前qps，如果qps>指定数值启动长连接
		if qps > 3 { //说明此时有3个冷启动的实例，大概率下面要处理高并发请求，所以需要建立长连接，启动一台实例
			fmt.Println("r.runRequest(ctx) 56")
			_, err := r.startNewRunner(ctx)
			if err != nil {
				return nil, err
			}
		} else {
			fmt.Println("r.runRequest(ctx) 62")
			return r.runRequest(ctx)
		}
	}
	runtimeRunner, exit := runners.GetOne()
	if !exit {
		panic("not running runner")
	}
	//fmt.Printf("Instance.Request %s\n", time.Now().String())
	runnerResponse, err := runtimeRunner.Request(ctx)
	if err != nil {
		return nil, err
	}
	return runnerResponse, nil

}
