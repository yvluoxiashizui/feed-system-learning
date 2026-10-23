package services

import (
	"strconv"
	"strings"

	"feed/models"
	"feed/repos"
)

// rankPage 榜单分页数据。游标是复合的：排序字段不唯一，必须同时带上排序值和 id
type rankPage struct {
	Videos     []models.Video `json:"videos"`
	NextCursor string         `json:"next_cursor"`
}

// encodeRankCursor 把「点赞数_视频 id」拼成游标
func encodeRankCursor(likes int64, id uint) string {
	return strconv.FormatInt(likes, 10) + "_" + strconv.FormatUint(uint64(id), 10)
}

// parseRankCursor 解析复合游标，格式不合法时返回零值，从头开始分页
func parseRankCursor(raw string) (int64, uint) {
	parts := strings.SplitN(raw, "_", 2)
	if len(parts) != 2 {
		return 0, 0
	}
	likes, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || likes < 0 {
		return 0, 0
	}
	id, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, 0
	}
	return likes, uint(id)
}

// LikesRank 点赞榜：按点赞数倒序，同分按 id 倒序。
// 游标带上点赞数和 id 两个字段，否则同分的那一批会在翻页时重复或漏掉。
func LikesRank(rawCursor string, limit int) (rankPage, error) {
	cursorLikes, cursorID := parseRankCursor(rawCursor)

	videos, err := repos.ListVideosByLikesCursor(cursorLikes, cursorID, limit)
	if err != nil {
		return rankPage{}, err
	}

	page := rankPage{Videos: videos}
	if len(videos) > 0 {
		last := videos[len(videos)-1]
		page.NextCursor = encodeRankCursor(last.LikesCount, last.ID)
	}
	return page, nil
}
