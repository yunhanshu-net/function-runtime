package v1

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/runcher/cmd"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/pkg/constants"
	"io"
)

func Runner(c *gin.Context) {

	req := request.RunnerRequest{
		Request: &request.Request{
			Method: c.Request.Method,
			Route:  c.Param("router"),
		},
		Runner: &model.Runner{
			Name:     c.Param("runner"),
			User:     c.Param("user"),
			Language: "go",
		},
	}
	if c.Request.Method == "GET" {
		req.Request.Body = c.Request.URL.RawQuery
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
		req.Request.Body = string(b)
		req.Request.BodyString = string(b)
	}

	traceID := c.GetHeader(constants.HttpTraceID)
	ctx := context.WithValue(context.Background(), constants.TraceID, traceID)
	get, err := cmd.Runcher.Scheduler.Request(ctx, &req)

	if err != nil {
		c.JSON(200, response.Body{
			Code: -1,
			Msg:  err.Error(),
		})
		fmt.Println(err)
		return
	}
	if v, ok := get.Body.(string); ok {
		c.Data(200, "application/json; charset=utf-8", []byte(v))
		return
	}
	fmt.Println("get", get)
	c.JSON(200, get.Body)
}
