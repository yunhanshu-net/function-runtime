package kernel

import (
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/kernel/coder"
	"github.com/yunhanshu-net/runcher/kernel/scheduler"
)

type Runcher struct {
	Scheduler  *scheduler.Scheduler
	Coder      *coder.Coder
	natsServer *server.Server
	natsConn   *nats.Conn
	down       chan struct{}
}

func MustNewRuncher() *Runcher {

	natsCli, natsSrv := InitNats()
	return &Runcher{
		down:       make(chan struct{}, 1),
		natsServer: natsSrv,
		natsConn:   natsCli,
		Scheduler:  scheduler.NewScheduler(natsCli),
		Coder:      coder.NewDefaultCoder(natsCli),
	}
}

func (a *Runcher) Run() error {
	go func() {
		err := a.Scheduler.Run()
		if err != nil {
			panic(err)
		}
	}()
	//err := a.Coder.Run()
	//if err != nil {
	//	return err
	//}

	//a.natsServer.WaitForShutdown()
	return nil
}
