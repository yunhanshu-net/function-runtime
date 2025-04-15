package kernel

import (
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"time"
)

var natsClient *nats.Conn
var natsServer *server.Server

func InitNats() (*nats.Conn, *server.Server) {

	opts := &server.Options{}

	// Initialize new server with options
	ns, err := server.NewServer(opts)

	if err != nil {
		panic(err)
	}

	// Start the server via goroutine
	go ns.Start()
	natsServer = ns

	// Wait for server to be ready for connections
	if !ns.ReadyForConnections(4 * time.Second) {
		panic("not ready for connection")
	}

	// Connect to server
	nc, err := nats.Connect(ns.ClientURL())

	if err != nil {
		panic(err)
	}
	natsClient = nc

	return nc, ns
	//if natsClient == nil {
	//	var err error
	//	natsClient, err = nats.Connect(nats.DefaultURL)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	//return natsClient
}
