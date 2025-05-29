package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/function-runtime/pkg/http2nats"
)

func Manage(c *gin.Context) {
	c.Set("trace_id", c.GetHeader("x-trace-id"))
	data, err := http2nats.GinRequest(c)
	if err != nil {
		c.JSON(500, map[string]interface{}{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}
	c.Data(200, "application/json", data)
}
