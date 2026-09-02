package handlers

import (
	"strconv"
	"net/http"
	"github.com/gin-gonic/gin"
	"feed/models"
)

func PublishVideo(c *gin.Context){
	userID := c.GetString("user_id")
	uid,_ := strconv.ParseUint(userID,10,64)

	var input struct {
		Title string	`json:"title"`
		PlayURL string	`json:"play_url"`
	}

	if err := c.ShouldBindJSON(&input);err != nil{
		c.JSON(400, gin.H{"error": "参数错误"})
		return 
	}

	var author models.User
	if err := db.Where("id = ?",uid).First(&author).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "查找错误"})
		return
	}
	video := models.Video{AuthorID: uint(uid), Username: author.Username, Title: input.Title, PlayURL: input.PlayURL}

	if err := db.Create(&video).Error ; err != nil{
		c.JSON(500, gin.H{"error": "查找错误"})
		return
	}
	
	c.JSON(http.StatusOK,video)
}

// ListVideos 视频列表（Feed流雏形）：按 id 倒序 = 最新在前
func ListVideos(c *gin.Context) {
	var videos []models.Video
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	l, _ := strconv.Atoi(limitStr)
	o, _ := strconv.Atoi(offsetStr) 
	db.Order("id DESC").Limit(l).Offset(o).Find(&videos)
	//limit
	c.JSON(200, videos)
}


