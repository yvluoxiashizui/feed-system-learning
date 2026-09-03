package handlers

import (
	"feed/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func FollowUser(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := strconv.ParseUint(userID, 10, 64)

	var input struct {
		VloggerID uint `json:"vlogger_id"`
	}
	c.ShouldBindJSON(&input)

	follow := models.Follow{
		FollowerID: uint(uid),
		VloggerID:  input.VloggerID,
	}

	if err := db.Create(&follow).Error; err != nil {
		c.JSON(400, gin.H{"error": "不能重复关注！"})
		return
	}

	c.JSON(200, gin.H{"message": "关注成功"})
}

func Unfollow(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := strconv.ParseUint(userID, 10, 64)

	var input struct {
		VloggerID uint `json:"vlogger_id"`
	}
	c.ShouldBindJSON(&input)
	
	if err := db.Where("follower_id = ? AND vlogger_id = ?",uint(uid),input.VloggerID).Delete(&models.Follow{}).Error; err != nil {
		c.JSON(400,gin.H{"error":"取关失败"})
		return
	}

	c.JSON(200,gin.H{"message":"已取关"})

}