package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	if tokenString == "" {
		c.JSON(401, gin.H{"error": "token不可为空"})
		c.Abort()
		return
	}
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte("feed-secret-key"), nil // 密钥，和登录签发时一致
	})

	if err != nil || !token.Valid {
		c.JSON(401, gin.H{"error": "token无效"})
		c.Abort()
		return
	}
	claims := token.Claims.(jwt.MapClaims)
	userID := fmt.Sprintf("%.0f", claims["user_id"].(float64))
	c.Set("user_id", userID)

	c.Next()
}
