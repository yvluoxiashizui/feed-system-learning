package handlers

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"

	"feed/services"
)

// 限流参数：rate 是每秒补充的令牌数，burst 是桶容量（允许的突发量）
const (
	RegisterRate  = 0.2 // 每 5 秒 1 次，突发 2 次
	RegisterBurst = 2
	LoginRate     = 2 // 每秒 2 次，突发 5 次
	LoginBurst    = 5
	PublishRate   = 0.5 // 每 2 秒 1 次，突发 3 次
	PublishBurst  = 3
)

// RateLimit 令牌桶限流中间件。未登录时按 IP 限流，登录后按用户 id 限流——
// 同一出口 IP 下的多个用户不会互相挤占，单个账号也不能靠换 IP 绕过。
func RateLimit(name string, rate float64, burst int) gin.HandlerFunc {
	retryAfter := strconv.Itoa(int(math.Ceil(1 / rate)))

	return func(c *gin.Context) {
		subject := c.ClientIP()
		if uid := c.GetString("user_id"); uid != "" {
			subject = "user:" + uid
		}

		allowed, err := services.AllowToken(name, subject, rate, burst)
		if err != nil {
			// 限流依赖的 Redis 不可用时放行，避免限流组件故障把整个接口拒掉
			c.Next()
			return
		}
		if !allowed {
			c.Header("Retry-After", retryAfter)
			c.AbortWithStatusJSON(429, gin.H{"error": "请求过于频繁，请稍后再试"})
			return
		}
		c.Next()
	}
}
