package v1

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/runcher/cmd"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
	"io"
)

func Http(c *gin.Context) {

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
		//mp := make(map[string]interface{})
		//queryMap := c.Request.URL.Query()
		//for k, values := range queryMap {
		//	if len(values) > 1 {
		//		mp[k] = values
		//	} else {
		//		mp[k] = values[0]
		//	}
		//}
		req.Request.Body = c.Request.URL.RawQuery
		//marshal, err := json.Marshal(mp)
		//if err != nil {
		//	panic(err)
		//}
		////req.Body = string(marshal)
		//req.Request.BodyString = string(marshal)

	}

	if c.Request.Method == "POST" {
		var r interface{}
		b, err := io.ReadAll(c.Request.Body)
		if err != nil {
			panic(err)
		}

		err = json.Unmarshal(b, &req.Request.Body)
		if err != nil {
			panic(err)
		}
		req.Request.Body = r
		req.Request.BodyString = string(b)
	}

	get, err := cmd.Runcher.Scheduler.Request(&req)

	if err != nil {
		c.JSON(200, nil)
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
