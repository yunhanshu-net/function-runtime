package response

type Response struct {
	MetaData     map[string]interface{} `json:"meta_data"` //内核层的元数据
	SoftResponse *SoftResponse          `json:"soft_response"`
}

type SoftResponse struct {
	MetaData   map[string]interface{} `json:"meta_data"` //SDK层元数据，例如日志，执行耗时，内存占用等等
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       *BizData               `json:"body"`
}

type BizData struct {
	TraceID  string                 `json:"trace_id"`
	MetaData map[string]interface{} `json:"meta_data"` //业务元数据，例如：投票的 bi 数据
	Msg      string                 `json:"msg"`
	Data     interface{}            `json:"data"`
	Code     int                    `json:"code"`
}

func (r *SoftResponse) JSON(statusCode int, data *BizData) error {
	r.StatusCode = statusCode
	r.Body = data
	return nil
}

func (r *SoftResponse) OKWithJSON(data interface{}, meta ...map[string]interface{}) error {
	bz := &BizData{Msg: "ok", Code: 0, Data: data}
	if len(meta) > 0 {
		bz.MetaData = meta[0]
	}
	return r.JSON(200, bz)
}
