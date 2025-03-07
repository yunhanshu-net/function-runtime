package kernel

import (
	"github.com/yunhanshu-net/runcher/runtime"
	"github.com/yunhanshu-net/runcher/transport"
	"sync"
)

type Executor struct {
	Transports map[string]transport.Transport //传输层，可能存在多种不同的实现，有nats，grpc，socket 套接字，等等
	Done       <-chan struct{}
	RWMutex    *sync.RWMutex
	Runners    map[string]map[int]*runtime.Runner //每个runner 可以有多个实例
}

func (r *Executor) GetRunners(subject string) (runners map[int]*runtime.Runner, exist bool) {
	r.RWMutex.RLock()
	defer r.RWMutex.RUnlock()
	runners, ok := r.Runners[subject]
	return runners, ok
}

func NewDefaultExecutor() (*Executor, error) {
	trs, err := transport.NewTransport(&transport.Config{TransportType: transport.TypeNats})
	if err != nil {
		return nil, err
	}
	return &Executor{RWMutex: &sync.RWMutex{}, Transports: map[string]transport.Transport{trs.GetConfig().TransportType: trs}}, nil
}

// Listen 这里主要是维护runner的状态信息
func (r *Executor) Listen() {
	for _, transportConn := range r.Transports {
		go func(transport transport.Transport) {
			err := transport.Connect()
			if err != nil {
				panic(err)
			}
			for {
				select {
				case msg, ok := <-transport.ReadMessage():
					if !ok {
						return
					}
					r.HandelRunnerStatus(msg)
				}
			}

		}(transportConn)
	}

	<-r.Done
	for _, transportConn := range r.Transports {
		transportConn.Close()
	}
}

func (r *Executor) GetTransport(transportType string) (transport.Transport, error) {
	return r.Transports[transportType], nil
}

func (r *Executor) Close() error {
	r.Done = make(<-chan struct{})
	return nil
}
