package transport

type MsgHeader map[string][]string

func (h MsgHeader) Add(key, value string) {
	h[key] = append(h[key], value)
}

// Set sets the header entries associated with key to the single
// element value. It is case-sensitive and replaces any existing
// values associated with key.
func (h MsgHeader) Set(key, value string) {
	h[key] = []string{value}
}

// Get gets the first value associated with the given key.
// It is case-sensitive.
func (h MsgHeader) Get(key string) string {
	if h == nil {
		return ""
	}
	if v := h[key]; v != nil {
		return v[0]
	}
	return ""
}

// Values returns all values associated with the given key.
// It is case-sensitive.
func (h MsgHeader) Values(key string) []string {
	return h[key]
}

// Del deletes the values associated with a key.
// It is case-sensitive.
func (h MsgHeader) Del(key string) {
	delete(h, key)
}

//func (b *Executor) RunnerListen() error {
//	connectKey := fmt.Sprintf("runner.connect.*.*")
//	closeKey := fmt.Sprintf("runner.close.*.*")
//	connectSub, err := b.nats.Subscribe(connectKey, func(msg *nats.Msg) {
//		fmt.Printf("connectKey:%s 建立长连接\n", connectKey)
//		b.RuncherConnectLock.Lock()
//		defer b.RuncherConnectLock.Unlock()
//		subject := msg.Subject
//		split := strings.Split(subject, ".")
//		if len(split) < 4 {
//			fmt.Println("RunnerListen: key err", split)
//			return
//		}
//		user := split[2]
//		runner := split[3]
//		if msg.Header.Get("status") == "0" {
//			h := nats.Header{}
//			h.Set("status", "0")
//			err := msg.RespondMsg(&nats.Msg{Header: h})
//			if err != nil {
//				fmt.Println("RunnerListen: RespondMsg key err", err.Error())
//				return
//			}
//			if _, ok := b.RuncherConnect[fmt.Sprintf("%s.%s", user, runner)]; !ok {
//				b.RuncherConnect[fmt.Sprintf("%s.%s", user, runner)] = make(chan ConnectStatus, 1)
//			}
//			b.RuncherConnect[fmt.Sprintf("%s.%s", user, runner)] <- ConnectStatus{Success: true}
//
//		}
//	})
//	if err != nil {
//		return err
//	}
//	closeSub, err := b.nats.Subscribe(closeKey, func(msg *nats.Msg) {
//		fmt.Printf("closeKey:%s 释放长连接\n", closeKey)
//
//		b.RuncherCloseLock.Lock()
//		defer b.RuncherCloseLock.Unlock()
//		subject := msg.Subject
//		split := strings.Split(subject, ".")
//		if len(split) < 4 {
//			fmt.Println("RunnerListen: key err", split)
//			return
//		}
//		user := split[2]
//		runner := split[3]
//		if msg.Header.Get("status") == "0" {
//			if _, ok := b.RuncherClose[fmt.Sprintf("%s.%s", user, runner)]; !ok {
//				b.RuncherClose[fmt.Sprintf("%s.%s", user, runner)] = make(chan CloseStatus, 1)
//			}
//			b.RuncherClose[fmt.Sprintf("%s.%s", user, runner)] <- CloseStatus{Success: true}
//		}
//	})
//	if err != nil {
//		return err
//	}
//	b.connectSub = connectSub
//	b.closeSub = closeSub
//	return nil
//}
//
//// WaitRunnerConnected 等待指定runner成功建立连接
//func (b *Executor) WaitRunnerConnected(ctx context.Context, user, app string) error {
//	select {
//	case <-ctx.Done():
//		return ctx.Err()
//	case <-time.After(time.Second * 5):
//		return fmt.Errorf("wait runner connected timeout")
//	case status := <-b.RuncherConnect[fmt.Sprintf("%s.%s", user, app)]:
//		if !status.Success {
//			return fmt.Errorf("wait runner connected err:" + status.Message)
//		}
//		return nil
//	}
//}
