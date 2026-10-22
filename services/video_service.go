package services

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

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

// 视频标题与播放地址的上限
const (
	maxVideoTitleLength = 100
	maxPlayURLLength    = 500
)

// PublishVideo 发布视频：标题和播放地址不能为空，并按字符数限制长度
func PublishVideo(userID uint, title, playURL string) (*models.Video, error) {
	title = strings.TrimSpace(title)
	playURL = strings.TrimSpace(playURL)

	if title == "" {
		return nil, errors.New("标题不能为空")
	}
	if utf8.RuneCountInString(title) > maxVideoTitleLength {
		return nil, errors.New("标题最多 100 字")
	}
	if playURL == "" {
		return nil, errors.New("播放地址不能为空")
	}
	if len(playURL) > maxPlayURLLength {
		return nil, errors.New("播放地址过长")
	}

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

// Feed 窗口缓存参数。窗口里存最近 N 条视频的 id（降序），
// 一个 key 覆盖前若干页，不同游标复用同一份数据。
const (
	feedWindowSize    = 500
	feedWindowKey     = "feed:recent:ids"
	feedCacheTTL      = 30 * time.Second
	emptyFeedCacheTTL = 10 * time.Second
)

// feedSingle 同一 key 并发未命中时只放一个请求回源
var feedSingle singleflight.Group

// ListFeed 游标分页 Feed。
// 优先走热点窗口缓存：窗口里能切出这一页就直接返回；切不出来（游标落在窗口之外，
// 或者窗口边界可能截断了真实结果）再回退到数据库，用 id 游标查。
func ListFeed(cursor uint, limit int) (feedPage, error) {
	if cacheEnabled() {
		if ids, err := recentFeedIDs(); err == nil {
			if page, ok := feedPageFromWindow(ids, cursor, limit); ok {
				return page, nil
			}
		}
	}
	return loadFeedPage(cursor, limit)
}

// feedPageFromWindow 从窗口里切出一页；切不出来时返回 false，交给调用方回源
func feedPageFromWindow(ids []uint, cursor uint, limit int) (feedPage, bool) {
	start := 0
	if cursor > 0 {
		start = -1
		for i, id := range ids {
			if id < cursor {
				start = i
				break
			}
		}
		if start < 0 {
			return feedPage{}, false
		}
	}

	end := start + limit
	if end > len(ids) {
		end = len(ids)
	}
	pageIDs := ids[start:end]

	// 没凑满又已经到窗口末尾，可能是窗口把真实结果截断了，回源保证正确
	if len(pageIDs) < limit && end == len(ids) {
		return feedPage{}, false
	}

	videos, err := repos.FindVideosByIDs(pageIDs)
	if err != nil {
		return feedPage{}, false
	}
	byID := make(map[uint]models.Video, len(videos))
	for _, v := range videos {
		byID[v.ID] = v
	}
	ordered := make([]models.Video, 0, len(pageIDs))
	for _, id := range pageIDs {
		if v, ok := byID[id]; ok {
			ordered = append(ordered, v)
		}
	}
	return feedPage{Videos: ordered, NextCursor: int(pageIDs[len(pageIDs)-1])}, true
}

// recentFeedIDs 取最近 feedWindowSize 条视频 id（降序），带缓存与单飞
func recentFeedIDs() ([]uint, error) {
	if ids, ok := readFeedWindow(); ok {
		return ids, nil
	}

	v, err, _ := feedSingle.Do(feedWindowKey, func() (interface{}, error) {
		// 等锁期间可能已被前一个请求回填，再查一次缓存
		if ids, ok := readFeedWindow(); ok {
			return ids, nil
		}
		ids, err := repos.ListRecentVideoIDs(feedWindowSize)
		if err != nil {
			return nil, err
		}
		writeFeedWindow(ids)
		return ids, nil
	})
	if err != nil {
		return nil, err
	}
	return v.([]uint), nil
}

// loadFeedPage 直接查库，并算出下一页游标（窗口覆盖不到时回退走这里）
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

// readFeedWindow 读窗口缓存，反序列化失败按未命中处理
func readFeedWindow() ([]uint, bool) {
	val, err := rdb.Get(ctx, feedWindowKey).Result()
	if err != nil {
		return nil, false
	}
	var ids []uint
	if err := json.Unmarshal([]byte(val), &ids); err != nil {
		return nil, false
	}
	return ids, true
}

// writeFeedWindow 写窗口缓存，过期时间加随机抖动，避免同一批写入的 key 同时失效
func writeFeedWindow(ids []uint) {
	data, err := json.Marshal(ids)
	if err != nil {
		return
	}
	base := feedCacheTTL
	if len(ids) == 0 {
		base = emptyFeedCacheTTL
	}
	jitter := time.Duration(rand.Int63n(int64(base / 5)))
	rdb.Set(ctx, feedWindowKey, data, base+jitter)
}

// GetVideoDetail 视频详情
func GetVideoDetail(id uint) (*models.Video, error) {
	return repos.FindVideoByID(id)
}

// LikeVideo 点赞：先确认视频存在，再在一个事务里插记录并累加计数，最后累加热度
func LikeVideo(userID, videoID uint) error {
	if _, err := repos.FindVideoByID(videoID); err != nil {
		return errors.New("视频不存在")
	}

	duplicated, err := repos.LikeAndIncrement(userID, videoID)
	if err != nil {
		return err
	}
	if duplicated {
		return errors.New("不能重复点赞")
	}
	rdb.ZIncrBy(ctx, "hot:videos", 1, strconv.Itoa(int(videoID))).Err()
	return nil
}

// UnlikeVideo 取消点赞：在一个事务里删记录并回退计数，再回退热度
func UnlikeVideo(userID, videoID uint) error {
	removed, err := repos.UnlikeAndDecrement(userID, videoID)
	if err != nil {
		return err
	}
	if !removed {
		return errors.New("还没有点赞")
	}

	member := strconv.Itoa(int(videoID))
	score, err := rdb.ZIncrBy(ctx, "hot:videos", -1, member).Result()
	if err == nil && score <= 0 {
		// 热度归零就从榜上摘掉，避免脏成员长期占位
		rdb.ZRem(ctx, "hot:videos", member)
	}
	return nil
}

// IsLiked 判断是否已点赞
func IsLiked(userID, videoID uint) (bool, error) {
	return repos.IsLiked(userID, videoID)
}

// LikedVideos 我赞过的视频，按点赞时间倒序，游标是点赞记录的 id
func LikedVideos(userID, cursor uint, limit int) (feedPage, error) {
	ids, next, err := repos.ListLikedVideoIDs(userID, cursor, limit)
	if err != nil {
		return feedPage{}, err
	}
	if len(ids) == 0 {
		return feedPage{Videos: []models.Video{}}, nil
	}

	videos, err := repos.FindVideosByIDs(ids)
	if err != nil {
		return feedPage{}, err
	}
	byID := make(map[uint]models.Video, len(videos))
	for _, v := range videos {
		byID[v.ID] = v
	}
	// 按点赞顺序重排，视频表的返回顺序不可依赖
	ordered := make([]models.Video, 0, len(ids))
	for _, id := range ids {
		if v, ok := byID[id]; ok {
			ordered = append(ordered, v)
		}
	}
	return feedPage{Videos: ordered, NextCursor: next}, nil
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
