package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/yunhanshu-net/runcher/api/v1"
	"github.com/yunhanshu-net/runcher/middleware"
)

func InitRouter(r *gin.Engine) {

	r.GET("/hello", func(c *gin.Context) {
		c.String(200, "ok")
	})

	api := r.Group("/api")
	api.Use(middleware.WithTraceId())
	api.Any("/runner/:user/:runner/*router", v1.Runner)

	//api.POST("/coder/AddApi", v1.AddApi)
	api.Any("/coder/:manage", v1.Manage)
	//api.POST("/coder/AddApis", v1.AddApis)
	//api.POST("/coder/AddBizPackage", v1.AddBizPackage)
	//api.POST("/coder/CreateProject", v1.CreateProject)
}
