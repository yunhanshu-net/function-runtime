package transport

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"sync"
)

type Config struct {
	TransportType string            `json:"transport_type"`
	Metadata      map[string]string `json:"metadata"`
	WorkPath      string            `json:"work_path"`
	RunnerType    string            `json:"runner_type"`
	Version       string            `json:"version"`
	Route         string            `json:"route"`  //命令
	User          string            `json:"user"`   //软件所属的用户
	Runner        string            `json:"runner"` //软件名
	OssPath       string            `json:"oss_path"`
	StartArgs     []string          `json:"start_args"`
}

type transportNats struct {
	wg               *sync.WaitGroup
	readMsgCount     int
	responseMsgCount int
	natsConn         *nats.Conn
	natsSub          *nats.Subscription
	msgList          chan *Msg
	status           *Status
	TransportConfig  *Config
}

func (t *transportNats) GetConfig() *Config {
	return t.TransportConfig
}

func (t *transportNats) GetStatus() *Status {
	return t.status
}

//函数请求
//runner.user.soft.version.run
//header 携带路由和
//函数请求

//关闭连接
//runner.user.soft.version.close 关闭连接请求
//关闭连接

//心跳检测，探针，判断调度引擎是否还存活正常
//runner.user.soft.version.heartbeat_check
//心跳检测，探针

func newTransportNats(transportConfig *Config) (trs *transportNats, err error) {
	return &transportNats{TransportConfig: transportConfig, wg: &sync.WaitGroup{}, msgList: make(chan *Msg, 10000), status: &Status{Status: "init"}}, nil

}

func (t *transportNats) ReadMessage() <-chan *Msg {
	return t.msgList
}

func (t *transportNats) Connect() error {

	url := t.TransportConfig.Metadata["nats-url"]
	if url == "" {
		url = nats.DefaultURL
	}
	group := t.TransportConfig.Metadata["nats-group"]
	if group == "" {
		group = fmt.Sprintf("%s.%s.%s", t.TransportConfig.User, t.TransportConfig.Runner, t.TransportConfig.Version)
	}

	conn, err := nats.Connect(url)
	if err != nil {
		return err
	}
	t.natsConn = conn

	//subject := fmt.Sprintf("runcher.%s.%s.%s.*", transportConfig.User, transportConfig.Runner, transportConfig.Version)
	//subject := fmt.Sprintf("runcher.%s.%s.%s.*", transportConfig.User, transportConfig.Runner, transportConfig.Version)
	//sub, err := conn.QueueSubscribe("runner.>", group, func(msg *nats.Msg) {
	sub, err := conn.QueueSubscribe("runcher.>", group, func(msg *nats.Msg) {
		t.wg.Add(1)
		t.readMsgCount++

		//fmt.Println("receive:", string(msg.Data))
		headers := make(map[string][]string)
		for k, v := range msg.Header {
			headers[k] = v
		}

		trMsg := &Msg{
			msg:       msg,
			Data:      msg.Data,
			Headers:   headers,
			Subject:   msg.Subject,
			transport: t,
		}
		t.msgList <- trMsg

	})
	if err != nil {
		return err
	}
	//trs = new(transportNats)
	t.natsSub = sub
	t.status.Status = "running"
	return nil

}

func (t *transportNats) Close() error {
	if t.status.Status == "closed" {
		return nil
	}
	t.status.Status = "closed"
	if t.natsConn != nil {
		t.natsConn.Close()
	}
	if t.natsSub != nil {
		t.natsSub.Unsubscribe()
	}
	close(t.msgList)
	return nil
}

func (t *transportNats) GetConn() interface{} {
	return t.natsConn
}
