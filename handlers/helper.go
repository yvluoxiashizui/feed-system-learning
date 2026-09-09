package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// currentUID 取当前登录用户 id（Auth 中间件存入上下文的 user_id）
func currentUID(c *gin.Context) uint {
	uid, _ := strconv.ParseUint(c.GetString("user_id"), 10, 64)
	return uint(uid)
}
