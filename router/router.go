package router

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/yunhanshu-net/function-runtime/api/v1"
	"github.com/yunhanshu-net/pkg/middleware"
	"time"
)

var start time.Time

func InitRouter(r *gin.Engine) {
	start = time.Now()

	r.GET("/hello", func(c *gin.Context) {
		c.String(200, "ok")
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Format(time.DateTime),
			"uptime":    time.Since(start).String(),
		})
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
