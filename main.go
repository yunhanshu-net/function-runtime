package main

import (
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/runcher/api"
	"github.com/yunhanshu-net/runcher/cmd"
)

func main() {
	//conns.Init()
	cmd.Init()
	//createtest.Create()

	v1 := gin.New()
	v1.Any("api/runner/:user/:runner/*router", api.Http)
	err2 := v1.Run(":9999")
	if err2 != nil {
		panic(err2)
	}

}
