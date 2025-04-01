package kernel

import (
	"github.com/yunhanshu-net/runcher/kernel/coder"
	"github.com/yunhanshu-net/runcher/kernel/scheduler"
)

type Runcher struct {
	Scheduler *scheduler.Scheduler
	Coder     *coder.Coder
}

func NewRuncher() *Runcher {
	conn := InitNats()
	return &Runcher{
		Scheduler: scheduler.NewScheduler(conn),
		Coder:     coder.NewDefaultCoder(conn),
	}
}

func (a *Runcher) Run() error {
	err := a.Scheduler.Run()
	if err != nil {
		return err
	}
	err = a.Coder.Run()
	if err != nil {
		return err
	}
	return nil
}
