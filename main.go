package main

import (
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/runcher/api"
	"github.com/yunhanshu-net/runcher/cmd"
)

func main() {
	//conns.Init()
	cmd.Init()
	//app := conns.Runcher
	//err2 := app.Run()
	//if err2 != nil {
	//	panic(err2)
	//}
	//
	//now := time.Now()

	//scheduler := v2.NewScheduler()
	//defer scheduler.Close()
	//req := &request.Request{
	//	Request: &request.RunnerRequest{
	//		UUID:   uuid.New().String(),
	//		Route:  "hello",
	//		Method: "GET",
	//		Body:   map[string]interface{}{"msg": "ok"}},
	//	Runner: &model.Runner{
	//		User:    "beiluo",
	//		Name:    "debug",
	//		Version: "v1",
	//	},
	//}

	//for i := 0; i < 10000; i++ {
	//	response, err := scheduler.Request(req)
	//	if err != nil {
	//		panic(err)
	//	}
	//	fmt.Println(jsonx.JSONString(response.Body))
	//}

	//var tasks []func()
	//for i := 0; i < 10000; i++ {
	//	tasks = append(tasks, func() {
	//		_, err := scheduler.Request(req)
	//		if err != nil {
	//			panic(err)
	//		}
	//		//fmt.Println(jsonx.JSONString(response.Body))
	//	})
	//}
	//syncx.ConcurrencyControl(tasks, 5)

	//response, err := scheduler.Request(req)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println(jsonx.JSONString(response.Body))
	//fmt.Printf("done cost:%s\n", time.Since(now).String())
	//r := &model.Runner{
	//	User:     "tencent",
	//	Version:  "v0",
	//	Name:     "openapi",
	//	Language: "go",
	//}
	//newRunner := runner.NewRunner(r)
	//err := newRunner.CreateProject()
	//if err != nil {
	//	panic(err)
	//}
	//pkg := &codex.BizPackage{
	//	AbsPackagePath: "array",
	//	Language:       "go",
	//	EnName:         "array",
	//	CnName:         "数组",
	//}
	//err2 = newRunner.AddBizPackage(pkg)
	//if err2 != nil {
	//	panic(err2)
	//}
	//cc := &codex.CodeApi{
	//	Package:        "array",
	//	AbsPackagePath: "array",
	//	//FilePath:       "diff",
	//	EnName: "diff",
	//	Code:   filex.LoadStringFromFile("/Users/yy/Desktop/code/github.com/runcher/temp/diff.go.temp"),
	//}
	//err2 = newRunner.AddApi(cc)
	//if err2 != nil {
	//	panic(err2)
	//}

	v1 := gin.New()
	v1.Any("api/runner/:user/:runner/*router", api.Http)
	err2 := v1.Run(":9999")
	if err2 != nil {
		panic(err2)
	}

}
