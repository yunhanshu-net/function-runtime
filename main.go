package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/api/v1"
	"github.com/yunhanshu-net/runcher/cmd"
)

func main() {
	//conns.Init()
	cmd.Init()
	//createtest.Create()

	app := gin.New()
	app.Any("/api/runner/:user/:runner/*router", v1.Http)
	app.GET("hello", func(c *gin.Context) {
		c.String(200, "ok")
	})
	//v1.Use(gin.Recovery()) // 添加 Recovery 中间件捕获 panic
	logrus.Info("start success")
	err2 := app.Run("0.0.0.0:9999")
	if err2 != nil {
		panic(err2)
	}

}
