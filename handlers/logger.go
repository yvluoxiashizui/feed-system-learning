package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func Logger(c *gin.Context) {
	fmt.Printf("收到请求：%v %v\n", c.Request.Method, c.Request.URL.Path)
	c.Next()
}
