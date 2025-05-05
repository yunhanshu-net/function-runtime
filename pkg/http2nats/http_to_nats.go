package http2nats

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"io"
	"strings"
	"time"
)

var conn *HttpToNats

func Setup(c *nats.Conn) {
	conn = &HttpToNats{
		c,
	}
}

type HttpToNats struct {
	//runcher *kernel.Runcher
	natsConn *nats.Conn
}

func GinRequest(c *gin.Context) (rspBodyData []byte, err error) {
	msg := &nats.Msg{}
	c.Request.URL.RequestURI()
	subject := strings.Split(c.Request.URL.RequestURI(), "?")[0]
	subject = strings.TrimPrefix(subject, "/api")
	subject = strings.Trim(subject, "/")
	subject = strings.ReplaceAll(subject, "/", ".")
	fmt.Println(subject)
	if c.Request.Method == "POST" {
		msg = nats.NewMsg(subject)
	} else {
		return nil, fmt.Errorf("invalid method")
	}
	traceId := c.GetHeader("x-trace-id")

	msg.Header.Set("trace_id", traceId)
	all, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	msg.Data = all
	natsConn := conn.natsConn
	rspMsg, err := natsConn.RequestMsg(msg, time.Second*20)
	if err != nil {
		return nil, err
	}
	code := rspMsg.Header.Get("code")
	if code != "0" {
		rspMsgStr := rspMsg.Header.Get("msg")
		return nil, fmt.Errorf("code: %s, msg: %s", code, rspMsgStr)
	}

	return rspMsg.Data, nil
}
