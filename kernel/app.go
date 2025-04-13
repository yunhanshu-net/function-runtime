package kernel

import (
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/kernel/coder"
	"github.com/yunhanshu-net/runcher/kernel/scheduler"
)

type Runcher struct {
	Scheduler *scheduler.Scheduler
	Coder     *coder.Coder
}

func MustNewRuncher() *Runcher {

	conn := InitNats()
	return &Runcher{
		Scheduler: scheduler.NewScheduler(),
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

	//a.natsServer.WaitForShutdown()
	return nil
}
