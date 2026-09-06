package repos

import "gorm.io/gorm"

// db 全局数据库连接，由 main 注入
var db *gorm.DB

// SetDB 注入数据库连接
func SetDB(d *gorm.DB) {
	db = d
}
