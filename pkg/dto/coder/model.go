package coder

type CodeApi struct {
	Language       string `json:"language"`
	Code           string `json:"code"`
	Package        string `json:"package"`
	AbsPackagePath string `json:"abs_package_path"` //这是绝对路径
	//FilePath       string `json:"file_path"`
	EnName string `json:"en_name"`
	CnName string `json:"cn_name"`
	Desc   string `json:"desc"`
}

type CodeApiCreateInfo struct {
	Language       string `json:"language"`
	Package        string `json:"package"`
	AbsPackagePath string `json:"abs_package_path"`
	//FilePath       string `json:"file_path"`
	EnName string `json:"en_name"`
	CnName string `json:"cn_name"`

	Msg    string `json:"msg"`
	Status string `json:"status"`
}

func (c *CodeApi) GetSubPackagePath() string {
	return c.AbsPackagePath
}

func (c *CodeApi) GetFileName() string {
	if c.Language == "" {
		c.Language = "go"
	}
	return c.EnName + "." + c.Language
}
