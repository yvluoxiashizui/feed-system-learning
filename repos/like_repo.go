package repos

import "feed/models"

// CreateLike 创建点赞记录，重复会因唯一索引报错
func CreateLike(l *models.Like) error {
	return db.Create(l).Error
}
