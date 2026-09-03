package models

import "time"

type Follow struct {
	ID uint `gorm:"primaryKey" json:"id"`
	FollowerID uint `gorm:"uniqueIndex:uk_follower_vlogger" json:"follower_id"`
	VloggerID uint `gorm:"uniqueIndex:uk_follower_vlogger" json:"vlogger_id"`
	CreatedAt time.Time `json:"created_at"`
}