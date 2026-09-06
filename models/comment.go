package models

import "time"

type Comment struct {
	ID uint `gorm:"primaryKey" json:"id"`
	VideoID uint `gorm:"index" json:"video_id"`
	UserID uint `json:"user_id"`
	Username string `json:"username"`
	Content string `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}