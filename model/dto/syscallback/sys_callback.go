package syscallback

import "encoding/json"

type Request[T any] struct {
	CallbackType string `json:"callback_type"`
	Data         T      `json:"data"`
}

func (s *Request[T]) DecodeData(el interface{}) error {
	marshal, err := json.Marshal(s.Data)
	if err != nil {
		return err
	}
	err = json.Unmarshal(marshal, &el)
	if err != nil {
		return err
	}
	return nil
}

type Response[T any] struct {
	Data T `json:"data"`
}
