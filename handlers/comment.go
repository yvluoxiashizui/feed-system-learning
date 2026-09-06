package handlers

import (
	"strconv"
	"feed/models"
	"github.com/gin-gonic/gin"
)

func PublishComment(c *gin.Context) {
	userID := c.GetString("user_id")
	uid,_ := strconv.ParseUint(userID ,10,64)

	var input struct {
		VideoID uint `json:"video_id"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&input);err != nil {
		c.JSON(400,gin.H{"error":"评论失败"})
		return
	}

	var author models.User
	if err := db.Where("id = ?",uid).First(&author).Error; err != nil {
		c.JSON(401, gin.H{"error": "用户不存在"})
		return
	}

	comment := models.Comment{
		UserID: uint(uid),
		Username: author.Username,
		VideoID:input.VideoID,
		Content:input.Content,
	}
	if err := db.Create(&comment).Error; err != nil {
		c.JSON(500, gin.H{"error": "发布失败"})
		return
	}
	c.JSON(200,comment)
}

func ListComments(c *gin.Context) {
	videoID, _ := strconv.ParseUint(c.Query("video_id"),10,64)

	var comments []models.Comment
	db.Where("video_id = ?",videoID).Order("id DESC").Find(&comments)
	c.JSON(200,comments)
}