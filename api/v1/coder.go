package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/yunhanshu-net/runcher/model/response"
	"github.com/yunhanshu-net/runcher/service/http2nats"
)

func Manage(c *gin.Context) {
	fmt.Println("Manage")
	data, err := http2nats.GinRequest(c)
	if err != nil {
		response.FailWithMessage(c, err.Error())
		return
	}
	c.Data(200, "application/json", data)
}
