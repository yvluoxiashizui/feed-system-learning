package models

import "time"

type Like struct{
	ID uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"uniqueIndex:uk_user_video" json:"user_id"`
	VideoID uint `gorm:"uniqueIndex:uk_user_video" json:"video_id"`
	CreatedAt time.Time `json:"created_at"`
}