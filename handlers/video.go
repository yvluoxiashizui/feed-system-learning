package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"feed/models"
)

// PublishVideo 发布视频（需登录）
func PublishVideo(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := strconv.ParseUint(userID, 10, 64)

	var input struct {
		Title   string `json:"title"`
		PlayURL string `json:"play_url"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 查作者，取用户名
	var author models.User
	if err := db.Where("id = ?", uid).First(&author).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}

	video := models.Video{
		AuthorID: uint(uid),
		Username: author.Username,
		Title:    input.Title,
		PlayURL:  input.PlayURL,
	}
	if err := db.Create(&video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "发布失败"})
		return
	}

	c.JSON(http.StatusOK, video)
}

// ListVideos Feed 流列表，分页，最新在前
func ListVideos(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var videos []models.Video
	db.Order("id DESC").Limit(limit).Offset(offset).Find(&videos)
	c.JSON(http.StatusOK, videos)
}
