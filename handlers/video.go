package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

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
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	// 查作者，取用户名
	var author models.User
	if err := db.Where("id = ?", uid).First(&author).Error; err != nil {
		c.JSON(401, gin.H{"error": "用户不存在"})
		return
	}

	video := models.Video{
		AuthorID: uint(uid),
		Username: author.Username,
		Title:    input.Title,
		PlayURL:  input.PlayURL,
	}
	if err := db.Create(&video).Error; err != nil {
		c.JSON(500, gin.H{"error": "发布失败"})
		return
	}

	c.JSON(200, video)
}

// ListVideos Feed 流列表，分页，最新在前
func ListVideos(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var videos []models.Video
	db.Order("id DESC").Limit(limit).Offset(offset).Find(&videos)
	c.JSON(200, videos)
}

// GetVideoDetail 视频详情
func GetVideoDetail(c *gin.Context) {
	id := c.Query("id")

	var video models.Video
	if err := db.Where("id = ?", id).First(&video).Error; err != nil {
		c.JSON(404, gin.H{"error": "视频不存在"})
		return
	}
	c.JSON(200, video)
}

// LikeVideo 点赞（需登录），同一用户不能重复点赞
func LikeVideo(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := strconv.ParseUint(userID, 10, 64)

	var input struct {
		VideoID uint `json:"video_id"`
	}
	c.ShouldBindJSON(&input)

	like := models.Like{
		UserID:uint(uid),
		VideoID: input.VideoID,
	}

	if err := db.Create(&like).Error;err != nil {
		c.JSON(400,gin.H{"error":"不能重复点赞"})
		return
	}

	db.Model(&models.Video{}).Where("id = ?", input.VideoID).UpdateColumn("likes_count", gorm.Expr("likes_count + 1"))
	c.JSON(200,gin.H{"message":"点赞成功"})

}
