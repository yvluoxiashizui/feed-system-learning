package handlers

import (
	"github.com/gin-gonic/gin"

	"feed/repos"
	"feed/services"
)

// Health 健康检查：数据库和 Redis 都可用才返回 200。
// Redis 只做缓存，挂掉时接口会自动回源，所以这里把它标成降级而不是不可用。
func Health(c *gin.Context) {
	if err := repos.Ping(); err != nil {
		c.JSON(503, gin.H{"status": "unavailable", "mysql": err.Error()})
		return
	}
	if err := services.PingRedis(); err != nil {
		c.JSON(200, gin.H{"status": "degraded", "redis": err.Error()})
		return
	}
	c.JSON(200, gin.H{"status": "ok"})
}
