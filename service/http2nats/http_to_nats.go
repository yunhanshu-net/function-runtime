package http2nats

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/yunhanshu-net/runcher/kernel"
	"io"
	"strings"
	"time"
)

var conn *HttpToNats

func Setup(runcher *kernel.Runcher) {
	conn = &HttpToNats{
		runcher,
	}
}

type HttpToNats struct {
	runcher *kernel.Runcher
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
	all, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	msg.Data = all
	natsConn := conn.runcher.GetNatsConn()
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
