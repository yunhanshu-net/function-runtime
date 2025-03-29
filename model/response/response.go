package response

type RunnerResponse struct {
	MetaData map[string]interface{} `json:"meta_data"` //内核层的元数据
	Response *Response              `json:"response"`
}

type Response struct {
	MetaData   map[string]interface{} `json:"meta_data"` //SDK层元数据，例如日志，执行耗时，内存占用等等
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       interface{}            `json:"body"`
}

//type BizData struct {
//	TraceID  string                 `json:"trace_id"`
//	MetaData map[string]interface{} `json:"meta_data"` //业务元数据，例如：投票的 bi 数据
//	Msg      string                 `json:"msg"`
//	Data     interface{}            `json:"data"`
//	Code     int                    `json:"code"`
//}
