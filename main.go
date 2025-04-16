package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yunhanshu-net/runcher/cmd"
	"github.com/yunhanshu-net/runcher/pkg/logger"
	"github.com/yunhanshu-net/runcher/router"
)

func main() {
	//conns.Init()
	logger.Setup()
	cmd.Init()
	cmd.Runcher.Run()

	app := gin.New()
	router.InitRouter(app)
	logrus.Info("start success")
	err2 := app.Run("0.0.0.0:9999")
	if err2 != nil {
		panic(err2)
	}

}
