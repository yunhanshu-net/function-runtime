package v2

import (
	"github.com/nats-io/nats.go"
)

type Upstream struct {
	conn        *nats.Conn
	upstreamSub *nats.Subscription
	manageSub   *nats.Subscription
}
