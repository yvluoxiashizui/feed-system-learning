package models

import "time"

// User 用户模型 = 数据库 users 表
// 你学过的：结构体 = 表，字段 = 列
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:32" json:"username"` // uniqueIndex = 唯一索引，用户名不能重复
	Password  string    `json:"-"`                                   // json:"-" = 密码绝不返回给前端
	CreatedAt time.Time `json:"created_at"`
}
