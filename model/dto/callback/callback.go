package callback

import "encoding/json"

type Request struct {
	Method string      `json:"method"`
	Router string      `json:"router"`
	Type   string      `json:"type"`
	Body   interface{} `json:"body"`
}

func (c *Request) DecodeData(el interface{}) error {
	marshal, err := json.Marshal(c.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(marshal, &el)
	if err != nil {
		return err
	}
	return nil
}

type ResponseWith[T, V any] struct {
	Request  T `json:"request"`
	Response V `json:"response"`
}
type Response struct {
	Request  interface{} `json:"request"`
	Response interface{} `json:"response"`
}
