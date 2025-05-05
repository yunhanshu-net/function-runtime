package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yunhanshu-net/runcher/pkg/constants"
	"time"
)

func WithTraceId() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(constants.HttpTraceID)
		if traceID == "" {
			u := time.Now().Format("20060102150405") + "-" + uuid.New().String()
			c.Request.Header.Set(constants.HttpTraceID, u)
			c.Set(constants.TraceID, u)
		}
		c.Next()
	}
}
