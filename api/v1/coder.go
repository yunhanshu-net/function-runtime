package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/pkg/http2nats"
)

func Manage(c *gin.Context) {
	c.Set("trace_id", c.GetHeader("x-trace-id"))
	data, err := http2nats.GinRequest(c)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	c.Data(200, "application/json", data)
}
