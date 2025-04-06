package response

type Response struct {
	MetaData map[string]interface{} `json:"meta_data"` //内核层的元数据
	Response *RunnerResponse        `json:"response"`
}

type Body struct {
	TraceID  string                 `json:"trace_id"`
	MetaData map[string]interface{} `json:"meta_data"` //sdk 层
	Code     int                    `json:"code"`
	Msg      string                 `json:"msg"`
	Data     interface{}            `json:"data"`
}

type RunnerResponse struct {
	MetaData   map[string]interface{} `json:"meta_data"` //SDK层元数据，例如日志，执行耗时，内存占用等等
	Headers    map[string]string      `json:"headers"`
	StatusCode int                    `json:"status_code"` //http对应http code 正常200
	Msg        string                 `json:"msg"`
	DataType   string                 `json:"data_type"`
	//Body       interface{}            `json:"body"`
	Body     Body `json:"body"`
	Multiple bool `json:"multiple"`
}

//type BizData struct {
//	TraceID  string                 `json:"trace_id"`
//	MetaData map[string]interface{} `json:"meta_data"` //业务元数据，例如：投票的 bi 数据
//	Msg      string                 `json:"msg"`
//	Data     interface{}            `json:"data"`
//	Code     int                    `json:"code"`
//}
