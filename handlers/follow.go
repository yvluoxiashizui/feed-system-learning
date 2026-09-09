package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"feed/services"
)

// FollowUser 关注用户（需登录）
func FollowUser(c *gin.Context) {
	uid := currentUID(c)

	var input struct {
		VloggerID uint `json:"vlogger_id"`
	}
	c.ShouldBindJSON(&input)

	if err := services.Follow(uint(uid), input.VloggerID); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "关注成功"})
}

// Unfollow 取消关注（需登录）
func Unfollow(c *gin.Context) {
	uid := currentUID(c)

	var input struct {
		VloggerID uint `json:"vlogger_id"`
	}
	c.ShouldBindJSON(&input)

	if err := services.Unfollow(uint(uid), input.VloggerID); err != nil {
		c.JSON(400, gin.H{"error": "取关失败"})
		return
	}
	c.JSON(200, gin.H{"message": "已取关"})
}

// IsFollowing 判断是否已关注（需登录）
func IsFollowing(c *gin.Context) {
	uid := currentUID(c)
	targetID, _ := strconv.ParseUint(c.Query("vlogger_id"), 10, 64)

	following, err := services.IsFollowing(uint(uid), uint(targetID))
	if err != nil {
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}
	c.JSON(200, gin.H{"is_following": following})
}

// FollowingFeed 关注流（需登录）
func FollowingFeed(c *gin.Context) {
	uid := currentUID(c)

	videos, err := services.FollowingFeed(uint(uid))
	if err != nil {
		c.JSON(500, gin.H{"error": "获取失败"})
		return
	}
	c.JSON(200, videos)
}
