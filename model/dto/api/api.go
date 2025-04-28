package api

type Params struct {

	//form,table,echarts,bi,3D .....
	RenderType string       `json:"render_type"`
	Children   []*ParamInfo `json:"children"`
}

type ParamInfo struct {
	//英文标识
	Code string `json:"code,omitempty"`
	//中文名称
	Name string `json:"name,omitempty"`
	//中文介绍
	Desc string `json:"desc,omitempty"`
	//是否必填
	Required bool `json:"required,omitempty"`

	Widget interface{} `json:"widget"`
}

type Info struct {
	Router      string   `json:"router"`
	Method      string   `json:"method"`
	User        string   `json:"user"`
	Runner      string   `json:"runner"`
	ApiDesc     string   `json:"api_desc"`
	Labels      []string `json:"labels"`
	ChineseName string   `json:"chinese_name"`
	EnglishName string   `json:"english_name"`
	Classify    string   `json:"classify"`
	//输入参数
	ParamsIn *Params `json:"params_in"`
	//输出参数
	ParamsOut *Params  `json:"params_out"`
	UseTables []string `json:"use_tables"`
	Callbacks []string `json:"callbacks"`
}

type ApiLogs struct {
	Version string  `json:"version"`
	Apis    []*Info `json:"apis"`
}
