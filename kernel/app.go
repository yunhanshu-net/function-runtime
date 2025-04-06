package kernel

import (
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/kernel/coder"
	v2 "github.com/yunhanshu-net/runcher/kernel/scheduler/v2"
)

type Runcher struct {
	Scheduler *v2.Scheduler
	Coder     *coder.Coder
}

func NewRuncher() *Runcher {
	conn := InitNats()
	return &Runcher{
		Scheduler: v2.NewScheduler(),
		Coder:     coder.NewDefaultCoder(conn),
	}
}

func (a *Runcher) Run() error {
	go func() {
		err := a.Scheduler.Run()
		logrus.Errorf(err.Error())
	}()
	err := a.Coder.Run()
	if err != nil {
		return err
	}

	return nil
}
