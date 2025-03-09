package transport

import "errors"

type Type string

func (t *Type) String() string {
	return string(*t)
}

const (
	TypeNats Type = "nats"
)

type Status struct {
	Status string `json:"status"` //running，closed
}

type Info interface {
	GetConfig() *Config
	GetConn() interface{}
	GetStatus() *Status
}

// Transport 通信层协议
type Transport interface {
	Info
	ReadMessage() <-chan *Msg
	Connect() error
	Close() error
}

func NewTransport(transportConfig *Config) (Transport, error) {
	if transportConfig == nil {
		return nil, errors.New("transport config is nil")
	}
	if transportConfig.TransportType == TypeNats || transportConfig.TransportType == "" {
		return newTransportNats(transportConfig)
	}

	panic("not support transport type")
}
