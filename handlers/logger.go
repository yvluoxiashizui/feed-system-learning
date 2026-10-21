package handlers

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger 请求日志中间件，记录方法、路径、状态码与耗时。
// 压测时设置 FEED_ACCESS_LOG=0 可关闭，避免 stdout 同步写成为吞吐瓶颈。
func Logger(c *gin.Context) {
	if os.Getenv("FEED_ACCESS_LOG") == "0" {
		c.Next()
		return
	}
	start := time.Now()
	c.Next()
	fmt.Printf("[%s] %s %d %s\n",
		c.Request.Method,
		c.Request.URL.Path,
		c.Writer.Status(),
		time.Since(start).Round(time.Microsecond),
	)
}
