package request

type Request struct {
	SoftInfo SoftInfo               `json:"soft_info"` //此次执行的软件信息
	TraceID  string                 `json:"trace_id"`  //分布式追踪
	Url      string                 `json:"url"`
	Method   string                 `json:"method"`
	Headers  map[string]string      `json:"headers"`
	MetaData map[string]interface{} `json:"meta_data"` //请求元数据
	Body     map[string]interface{} `json:"body"`      //请求json
	FileMap  map[string][]string    `json:"file_map"`
}

type SoftInfo struct {
	RequestJsonPath string `json:"request_json_path"` //请求参数存储路径
	WorkPath        string `json:"work_path"`         //执行目录
	RunnerType      string `json:"runner_type"`       //软件类型
	User            string `json:"user"`              //软件所属的用户
	Soft            string `json:"soft"`              //软件名
	Command         string `json:"command"`           //命令
	Version         string `json:"version"`           //版本
	SavePath        string `json:"save_path"`         //软件存储的 oss 地址
}
