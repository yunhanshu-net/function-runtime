package runner

import (
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/transport"
)

type Context struct {
	Transport transport.Info

	Conn   *nats.Conn
	Status string
}

func (c *Context) IsRunning() bool {
	return c.Status == "running"
}
