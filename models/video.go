package models

import "time"

// Video 视频表
type Video struct {
	// 主键同时作为点赞榜联合索引的第二列，用于同分时按 id 倒序定位
	ID          uint      `gorm:"primaryKey;index:idx_likes_count_id,priority:2" json:"id"`
	AuthorID    uint      `gorm:"index" json:"author_id"`
	Username    string    `json:"username"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	PlayURL     string    `json:"play_url"`
	CoverURL    string    `json:"cover_url"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"create_time"`
	// 点赞榜按 (likes_count, id) 排序，联合索引让这条排序直接走索引，不需要额外排序
	LikesCount int64 `gorm:"default:0;index:idx_likes_count_id,priority:1" json:"likes_count"`
}
