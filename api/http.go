package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/runcher/cmd"
	"github.com/yunhanshu-net/runcher/model"
	"github.com/yunhanshu-net/runcher/model/request"
)

func Http(c *gin.Context) {

	req := request.Request{
		Request: &request.RunnerRequest{
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
		mp := make(map[string]interface{})
		queryMap := c.Request.URL.Query()
		for k, values := range queryMap {
			if len(values) > 1 {
				mp[k] = values
			} else {
				mp[k] = values[0]
			}
		}
		req.Request.Body = mp

	}

	if c.Request.Method == "POST" {
		var r interface{}
		err := c.ShouldBindJSON(&r)
		if err != nil {
			panic(err)
		}
		req.Request.Body = r
	}

	get, err := cmd.Runcher.Scheduler.Request(&req)
	if err != nil {
		c.JSON(200, nil)
		fmt.Println(err)
		return
	}
	c.JSON(200, get.Body)
}
