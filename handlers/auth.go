package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Auth JWT 鉴权中间件，验证通过才放行
func Auth(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(401, gin.H{"error": "未登录"})
		c.Abort()
		return
	}

	// 解析并验证签名
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte("feed-secret-key"), nil
	})
	if err != nil || !token.Valid {
		c.JSON(401, gin.H{"error": "token 无效"})
		c.Abort()
		return
	}

	// 用户 id 存入上下文，供后续 handler 使用
	claims := token.Claims.(jwt.MapClaims)
	userID := fmt.Sprintf("%.0f", claims["user_id"].(float64))
	c.Set("user_id", userID)

	c.Next()
}
