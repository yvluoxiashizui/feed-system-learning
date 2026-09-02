package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Logger 请求日志中间件
func Logger(c *gin.Context) {
	fmt.Printf("[%s] %s\n", c.Request.Method, c.Request.URL.Path)
	c.Next()
}
