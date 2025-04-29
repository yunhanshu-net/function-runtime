package response

import (
	"encoding/json"
	"net/http"
)

type Body struct {
	DataType string                 `json:"data_type"`
	TraceID  string                 `json:"trace_id"`
	MetaData map[string]interface{} `json:"meta_data"` //sdk 层
	Code     int                    `json:"code"`
	Msg      string                 `json:"msg"`
	Data     interface{}            `json:"data"`
}

type BodyWith[T any] struct {
	DataType string                 `json:"data_type"`
	TraceID  string                 `json:"trace_id"`
	MetaData map[string]interface{} `json:"meta_data"` //sdk 层
	Code     int                    `json:"code"`
	Msg      string                 `json:"msg"`
	Data     T                      `json:"data"`
}

func (b *Body) DecodeData() {

}

type Response struct {
	MetaData   map[string]interface{} `json:"meta_data"` //SDK层元数据，例如日志，执行耗时，内存占用等等
	Headers    map[string]string      `json:"headers"`
	StatusCode int                    `json:"status_code"` //http对应http code 正常200
	Msg        string                 `json:"msg"`
	DataType   string                 `json:"data_type"`
	Body       interface{}            `json:"body"`
	Multiple   bool                   `json:"multiple"`
}

func (r *Response) OK() bool {
	return r.StatusCode == http.StatusOK
}

type With[T any] struct {
	MetaData   map[string]interface{} `json:"meta_data"` //SDK层元数据，例如日志，执行耗时，内存占用等等
	Headers    map[string]string      `json:"headers"`
	StatusCode int                    `json:"status_code"` //http对应http code 正常200
	Msg        string                 `json:"msg"`
	DataType   string                 `json:"data_type"`
	Body       T                      `json:"body"`
	Multiple   bool                   `json:"multiple"`
}

func (r *Response) DecodeBody() (*Body, error) {
	var bd Body
	switch r.Body.(type) {
	case string:
		err := json.Unmarshal([]byte(r.Body.(string)), &bd)
		if err != nil {
			return nil, err
		}
	default:
		marshal, err := json.Marshal(r.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(marshal, &bd)
		if err != nil {
			return nil, err
		}
	}
	return &bd, nil
}
func DecodeBody(r *Response, body interface{}) error {

	switch r.Body.(type) {
	case string:
		err := json.Unmarshal([]byte(r.Body.(string)), body)
		if err != nil {
			return err
		}
	default:
		marshal, err := json.Marshal(r.Body)
		if err != nil {
			return err
		}
		err = json.Unmarshal(marshal, body)
		if err != nil {
			return err
		}
	}
	return nil
}
