package v1

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/function-go/pkg/dto/request"
	"github.com/yunhanshu-net/function-go/pkg/dto/response"
	"github.com/yunhanshu-net/function-runtime/cmd"
	"github.com/yunhanshu-net/function-runtime/conf"
	"github.com/yunhanshu-net/pkg/constants"
	"github.com/yunhanshu-net/pkg/dto/runnerproject"
	"io"
)

func Runner(c *gin.Context) {
	traceID := c.GetHeader(constants.HttpTraceID)
	runner, err := runnerproject.NewRunner(c.Param("user"), c.Param("runner"), conf.GetRunnerRoot())
	if err != nil {
		panic(err)
	}
	req := request.RunFunctionReq{
		TraceID: traceID,
		Method:  c.Request.Method,
		Router:  c.Param("router"),
		Runner:  runner,
	}
	if c.Request.Method == "GET" {
		req.UrlQuery = c.Request.URL.RawQuery
	}

	if c.Request.Method == "POST" {
		//var r interface{}
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			panic(err)
		}
		defer c.Request.Body.Close()

		//err = json.Unmarshal(b, &req.Request.Body)
		//if err != nil {
		//	panic(err)
		//}
		fmt.Println(string(b))
		req.Body = string(b)
		req.BodyType = "string"
	}

	ctx := context.WithValue(context.Background(), constants.TraceID, traceID)
	get, err := cmd.Runcher.Scheduler.Request(ctx, &req)

	if err != nil {
		c.JSON(200, response.RunFunctionResp{Code: -1, Msg: err.Error()})
		fmt.Println(err)
		return
	}
	//if v, ok := get.Body.(string); ok {
	//	c.Data(200, "application/json; charset=utf-8", []byte(v))
	//	return
	//}
	fmt.Println("get", get)
	c.JSON(200, get)
}
