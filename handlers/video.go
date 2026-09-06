package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

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

// feedResponse Feed 流分页响应
type feedResponse struct {
	Videos     []models.Video `json:"videos"`
	NextCursor int            `json:"next_cursor"`
}

// ListVideos Feed 流列表，游标分页，最新在前
func ListVideos(c *gin.Context) {
	ctx := context.Background()
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	cursor, _ := strconv.ParseUint(c.DefaultQuery("cursor", "0"), 10, 64)

	// 缓存 key 带上 cursor，不同页不同缓存
	key := fmt.Sprintf("feed:videos:%d:%d", cursor, limit)

	// 查 Redis 缓存
	val, err := rdb.Get(ctx, key).Result()
	if err == nil {
		var cached feedResponse
		json.Unmarshal([]byte(val), &cached)
		c.JSON(200, cached)
		return
	}

	// 没命中查 MySQL（游标：取比 cursor 更早的视频）
	var videos []models.Video
	query := db.Order("id DESC").Limit(limit)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	query.Find(&videos)

	// 组装响应：next_cursor = 本页最后一个视频的 id
	resp := feedResponse{Videos: videos}
	if len(videos) > 0 {
		resp.NextCursor = int(videos[len(videos)-1].ID)
	}

	// 写缓存
	data, _ := json.Marshal(resp)
	rdb.Set(ctx, key, data, 30*time.Second)

	c.JSON(200, resp)
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
		UserID:  uint(uid),
		VideoID: input.VideoID,
	}

	if err := db.Create(&like).Error; err != nil {
		c.JSON(400, gin.H{"error": "不能重复点赞"})
		return
	}

	ctx := context.Background()

	db.Model(&models.Video{}).Where("id = ?", input.VideoID).UpdateColumn("likes_count", gorm.Expr("likes_count + 1"))
	rdb.ZIncrBy(ctx, "hot:videos", 1, strconv.Itoa(int(input.VideoID))).Err()
	c.JSON(200, gin.H{"message": "点赞成功"})
}

// 热榜接口
func HotVideos(c *gin.Context) {
	ctx := context.Background()

	ids, _ := rdb.ZRevRange(ctx, "hot:videos", 0, 9).Result()
	if len(ids) == 0 {
		c.JSON(200, []models.Video{})
		return
	}

	var videoIDs []uint
	for _, id := range ids {
		n, _ := strconv.ParseUint(id, 10, 64)
		videoIDs = append(videoIDs, uint(n))
	}

	var videos []models.Video
	db.Where("id IN ?", videoIDs).Find(&videos)

	//存map,按videoIDs顺序重排
	videoMap := map[uint]models.Video{}
	for _, v := range videos {
		videoMap[v.ID] = v
	}

	result := []models.Video{}
	for _, id := range videoIDs {
		if v, ok := videoMap[id]; ok {
			result = append(result, v)
		}
	}
	c.JSON(200, result)
}
