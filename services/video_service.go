package services

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"feed/models"
	"feed/repos"
)

var ctx = context.Background()

// feedPage Feed 流分页数据（缓存用）
type feedPage struct {
	Videos     []models.Video `json:"videos"`
	NextCursor int            `json:"next_cursor"`
}

// PublishVideo 发布视频（查作者、建记录）
func PublishVideo(userID uint, title, playURL string) (*models.Video, error) {
	author, err := repos.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	v := &models.Video{AuthorID: author.ID, Username: author.Username, Title: title, PlayURL: playURL}
	if err := repos.CreateVideo(v); err != nil {
		return nil, err
	}
	return v, nil
}

// ListFeed 游标分页 Feed，带 Redis 缓存
func ListFeed(cursor uint, limit int) (feedPage, error) {
	key := "feed:videos:" + strconv.Itoa(int(cursor)) + ":" + strconv.Itoa(limit)

	// 查缓存
	if val, err := rdb.Get(ctx, key).Result(); err == nil {
		var cached feedPage
		json.Unmarshal([]byte(val), &cached)
		return cached, nil
	}

	// 查库
	videos, err := repos.ListVideosByCursor(cursor, limit)
	if err != nil {
		return feedPage{}, err
	}
	page := feedPage{Videos: videos}
	if len(videos) > 0 {
		page.NextCursor = int(videos[len(videos)-1].ID)
	}

	// 写缓存
	data, _ := json.Marshal(page)
	rdb.Set(ctx, key, data, 30*time.Second)
	return page, nil
}

// GetVideoDetail 视频详情
func GetVideoDetail(id uint) (*models.Video, error) {
	return repos.FindVideoByID(id)
}

// LikeVideo 点赞：插记录 + 计数自增 + 热度累计
func LikeVideo(userID, videoID uint) error {
	like := &models.Like{UserID: userID, VideoID: videoID}
	if err := repos.CreateLike(like); err != nil {
		return errors.New("不能重复点赞")
	}
	repos.IncrementVideoLikes(videoID)
	rdb.ZIncrBy(ctx, "hot:videos", 1, strconv.Itoa(int(videoID))).Err()
	return nil
}

// HotVideos 热度 Top10
func HotVideos() ([]models.Video, error) {
	ids, _ := rdb.ZRevRange(ctx, "hot:videos", 0, 9).Result()
	if len(ids) == 0 {
		return []models.Video{}, nil
	}
	var videoIDs []uint
	for _, id := range ids {
		n, _ := strconv.ParseUint(id, 10, 64)
		videoIDs = append(videoIDs, uint(n))
	}
	videos, err := repos.FindVideosByIDs(videoIDs)
	if err != nil {
		return nil, err
	}
	// 按热度顺序重排
	videoMap := map[uint]models.Video{}
	for _, v := range videos {
		videoMap[v.ID] = v
	}
	result := make([]models.Video, 0, len(videoIDs))
	for _, id := range videoIDs {
		if v, ok := videoMap[id]; ok {
			result = append(result, v)
		}
	}
	return result, nil
}
