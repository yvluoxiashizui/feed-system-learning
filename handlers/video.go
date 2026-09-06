package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"feed/services"
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

	video, err := services.PublishVideo(uint(uid), input.Title, input.PlayURL)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, video)
}

// ListVideos Feed 流列表，游标分页
func ListVideos(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	cursor, _ := strconv.ParseUint(c.DefaultQuery("cursor", "0"), 10, 64)

	page, err := services.ListFeed(uint(cursor), limit)
	if err != nil {
		c.JSON(500, gin.H{"error": "获取失败"})
		return
	}
	c.JSON(200, page)
}

// GetVideoDetail 视频详情
func GetVideoDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Query("id"), 10, 64)

	video, err := services.GetVideoDetail(uint(id))
	if err != nil {
		c.JSON(404, gin.H{"error": "视频不存在"})
		return
	}
	c.JSON(200, video)
}

// LikeVideo 点赞（需登录）
func LikeVideo(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := strconv.ParseUint(userID, 10, 64)

	var input struct {
		VideoID uint `json:"video_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if err := services.LikeVideo(uint(uid), input.VideoID); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "点赞成功"})
}

// HotVideos 热榜 Top10
func HotVideos(c *gin.Context) {
	videos, err := services.HotVideos()
	if err != nil {
		c.JSON(500, gin.H{"error": "获取失败"})
		return
	}
	c.JSON(200, videos)
}
