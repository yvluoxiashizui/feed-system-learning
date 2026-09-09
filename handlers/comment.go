package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"feed/services"
)

// PublishComment 发表评论（需登录）
func PublishComment(c *gin.Context) {
	uid := currentUID(c)

	var input struct {
		VideoID uint   `json:"video_id"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	comment, err := services.PublishComment(uint(uid), input.VideoID, input.Content)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, comment)
}

// ListComments 评论列表
func ListComments(c *gin.Context) {
	videoID, _ := strconv.ParseUint(c.Query("video_id"), 10, 64)

	comments, err := services.ListComments(uint(videoID))
	if err != nil {
		c.JSON(500, gin.H{"error": "获取失败"})
		return
	}
	c.JSON(200, comments)
}
