package runner

import "github.com/yunhanshu-net/runcher/transport"

type Context struct {
	Transport transport.Info
	Status    string
}

func (c *Context) IsRunning() bool {
	return c.Status == "running"
}
