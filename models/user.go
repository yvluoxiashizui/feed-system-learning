package models

import "time"

// User 用户表
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:32" json:"username"` // 用户名唯一
	Password  string    `json:"-"`                                   // 密码加密存储，不返回给前端
	CreatedAt time.Time `json:"created_at"`
}
