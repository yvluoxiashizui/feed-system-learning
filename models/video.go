package models

import "time"

// Video 视频表
type Video struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	AuthorID    uint      `gorm:"index" json:"author_id"`
	Username    string    `json:"username"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	PlayURL     string    `json:"play_url"`
	CoverURL    string    `json:"cover_url"`
	CreateTime  time.Time `gorm:"autoCreateTime" json:"create_time"`
	LikesCount  int64     `gorm:"default:0" json:"likes_count"`
}
