package services

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"strconv"
	"time"

	"golang.org/x/sync/singleflight"

	"feed/models"
	"feed/repos"
)

var ctx = context.Background()

// cacheEnabled Feed 缓存开关，默认开启。
// 压测时设置 FEED_CACHE_DISABLED=1 可关闭缓存读取与写入，
// 用于在同一份代码上做「有缓存 / 无缓存」的读写路径对照。
func cacheEnabled() bool {
	return os.Getenv("FEED_CACHE_DISABLED") != "1"
}

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

// Feed 缓存过期时间基准。空结果取更短的值，避免新视频发布后被空值挡住。
const (
	feedCacheTTL      = 30 * time.Second
	emptyFeedCacheTTL = 10 * time.Second
)

// feedSingle 同一 key 并发未命中时只放一个请求回源
var feedSingle singleflight.Group

// ListFeed 游标分页 Feed，带 Redis 缓存
func ListFeed(cursor uint, limit int) (feedPage, error) {
	key := feedCacheKey(cursor, limit)

	// 查缓存；反序列化失败按未命中处理，回源查库
	if cacheEnabled() {
		if page, ok := readFeedCache(key); ok {
			return page, nil
		}
	}

	// 未命中时收敛：同一 key 同时只让一个请求查库，其余等它的结果
	v, err, _ := feedSingle.Do(key, func() (interface{}, error) {
		// 等锁期间可能已被前一个请求回填，再查一次缓存
		if cacheEnabled() {
			if page, ok := readFeedCache(key); ok {
				return page, nil
			}
		}
		page, err := loadFeedPage(cursor, limit)
		if err != nil {
			return feedPage{}, err
		}
		if cacheEnabled() {
			writeFeedCache(key, page)
		}
		return page, nil
	})
	if err != nil {
		return feedPage{}, err
	}
	return v.(feedPage), nil
}

func feedCacheKey(cursor uint, limit int) string {
	return "feed:videos:" + strconv.Itoa(int(cursor)) + ":" + strconv.Itoa(limit)
}

// loadFeedPage 直接查库，并算出下一页游标
func loadFeedPage(cursor uint, limit int) (feedPage, error) {
	videos, err := repos.ListVideosByCursor(cursor, limit)
	if err != nil {
		return feedPage{}, err
	}
	page := feedPage{Videos: videos}
	if len(videos) > 0 {
		page.NextCursor = int(videos[len(videos)-1].ID)
	}
	return page, nil
}

// readFeedCache 读缓存，反序列化失败按未命中处理
func readFeedCache(key string) (feedPage, bool) {
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return feedPage{}, false
	}
	var page feedPage
	if err := json.Unmarshal([]byte(val), &page); err != nil {
		return feedPage{}, false
	}
	return page, true
}

// writeFeedCache 写缓存，过期时间加随机抖动，避免同一批写入的 key 同时失效
func writeFeedCache(key string, page feedPage) {
	data, err := json.Marshal(page)
	if err != nil {
		return
	}
	base := feedCacheTTL
	if len(page.Videos) == 0 {
		base = emptyFeedCacheTTL
	}
	jitter := time.Duration(rand.Int63n(int64(base / 5)))
	rdb.Set(ctx, key, data, base+jitter)
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
