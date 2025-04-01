package kernel

import "github.com/nats-io/nats.go"

var natsClient *nats.Conn

func InitNats() *nats.Conn {
	if natsClient == nil {
		var err error
		natsClient, err = nats.Connect(nats.DefaultURL)
		if err != nil {
			panic(err)
		}
	}
	return natsClient
}
