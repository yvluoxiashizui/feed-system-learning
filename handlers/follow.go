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

func IsFollowing(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := strconv.ParseUint(userID, 10, 64)
	targetID,_ := strconv.ParseUint(c.Query("vlogger_id"),10,64)

	var count int64
	db.Model(&models.Follow{}).
		Where("follower_id = ? AND vlogger_id = ?",uid,targetID).
		Count(&count)

	isFollowing := count > 0

	c.JSON(200,gin.H{"is_following":isFollowing})
}

func FollowingFeed(c *gin.Context) {
	userID := c.GetString("user_id")
	uid, _ := strconv.ParseUint(userID, 10, 64)

	var follows []models.Follow
	db.Where("follower_id = ?",uid).Find(&follows)

	var vloggerIDs []uint
	for _,f := range follows {
		vloggerIDs = append(vloggerIDs, f.VloggerID)
	}

	var videos []models.Video
	if len(vloggerIDs) == 0 {
      c.JSON(200, []models.Video{})   // 返回空列表
      return
	}
	db.Where("author_id IN ?",vloggerIDs).Order("id DESC").Find(&videos)

	c.JSON(200,videos)
}