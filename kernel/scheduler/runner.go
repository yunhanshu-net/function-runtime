package scheduler

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/runner"
	"github.com/yunhanshu-net/runcher/runtime"
	"time"
)

type runnerReady struct {
	UUID   string
	Runner *runtime.Runner
	Err    error
}

func (r *Scheduler) getRunner(ctx *request.Request) (runners *runtime.Runners, exist bool) {
	r.lk.Lock()
	runners, ok := r.runners[ctx.GetSubject()]
	r.lk.Unlock()
	if !ok {
		return nil, false
	}
	return runners, true
}

func (r *Scheduler) addRunningRunner(ctx *request.Request) (runners *runtime.Runners, exist bool) {
	r.lk.Lock()
	runners, ok := r.runners[ctx.GetSubject()]
	r.lk.Unlock()
	if !ok {
		return nil, false
	}
	return runners, true
}

// 长连接
func (r *Scheduler) startNewRunner(reqCtx *request.Context) (runner.Runner, error) {
	newRunner := runner.NewRunner(reqCtx.Request.Runner)
	subject := reqCtx.Request.GetSubject()
	uid := uuid.New().String()
	fmt.Printf("start:%s\n", uid)
	reqCtx.Request.UUID = uid
	r.lk.Lock()
	ready := make(chan runnerReady, 1)
	runtimeRunner := &runtime.Runner{
		Conn:      r.conn,
		UUID:      uid,
		StartTime: time.Now(),
		Instance:  newRunner,
		Status:    "running",
	}
	wait := &waitReady{
		ready:  ready,
		runner: runtimeRunner,
	}
	r.waitRunnerReady[uid] = wait
	r.lk.Unlock()
	now := time.Now()
	err := newRunner.StartKeepAlive(reqCtx)
	if err != nil {
		return nil, err
	}
	select {
	case <-ready:
		r.runners[subject].Running = append(r.runners[subject].Running, runtimeRunner)
		sub := time.Now().Sub(now)
		fmt.Printf("建立连接总计耗时：%s\n", sub.String())
		return newRunner, nil
	case <-time.After(time.Second * 10):
		return nil, fmt.Errorf("startNewRunner timeout")
	}

}

// 临时执行，即刻释放
func (r *Scheduler) execRunner(reqCtx *request.Context) (*response.Response, error) {
	return nil, nil
}

func (r *Scheduler) runRequest(reqCtx *request.Context) (*response.Response, error) {
	newRunner := runner.NewRunner(reqCtx.Request.Runner)
	runnerResponse, err := newRunner.Request(reqCtx)
	if err != nil {
		return nil, err
	}
	return runnerResponse, nil
}

func (r *Scheduler) removeRunner(subject string, uuid string) {
	//r.lk.Lock()
	//defer r.lk.Unlock()
	runners := r.runners[subject]
	runners.RemoveRunner(uuid)
}
