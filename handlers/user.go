package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"feed/services"
)

// Register 注册
func Register(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}
	user, err := services.Register(input.Username, input.Password)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"id": user.ID, "username": user.Username})
}

// Login 登录
func Login(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}
	token, user, err := services.Login(input.Username, input.Password)
	if err != nil {
		c.JSON(401, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"token": token, "user_id": user.ID, "username": user.Username})
}

// GetUser 按 ID 查用户
func GetUser(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)

	user, err := services.GetUser(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}
	c.JSON(200, user)
}
