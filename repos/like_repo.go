package repos

import (
	"gorm.io/gorm"

	"feed/models"
)

// CreateLike 创建点赞记录，重复会因唯一索引报错
func CreateLike(l *models.Like) error {
	return db.Create(l).Error
}

// UnlikeAndDecrement 在一个事务内删除点赞记录并回退计数。
// 返回 removed 表示确实删掉了一条记录；记录不存在时不动计数，
// 避免重复取消把点赞数减成负数。
func UnlikeAndDecrement(userID, videoID uint) (bool, error) {
	removed := false
	err := db.Transaction(func(tx *gorm.DB) error {
		res := tx.Where("user_id = ? AND video_id = ?", userID, videoID).Delete(&models.Like{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		removed = true
		// 条件里带 likes_count > 0 兜底，历史脏数据也不会被减成负数
		return tx.Model(&models.Video{}).
			Where("id = ? AND likes_count > 0", videoID).
			UpdateColumn("likes_count", gorm.Expr("likes_count - 1")).Error
	})
	if err != nil {
		return false, err
	}
	return removed, nil
}

// IsLiked 判断用户是否已点赞
func IsLiked(userID, videoID uint) (bool, error) {
	var count int64
	if err := db.Model(&models.Like{}).
		Where("user_id = ? AND video_id = ?", userID, videoID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListLikedVideoIDs 按点赞记录 id 倒序取用户赞过的视频 id，并返回下一页游标
func ListLikedVideoIDs(userID, cursor uint, limit int) ([]uint, int, error) {
	var likes []models.Like
	query := db.Where("user_id = ?", userID).Order("id DESC").Limit(limit)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	if err := query.Find(&likes).Error; err != nil {
		return nil, 0, err
	}

	ids := make([]uint, 0, len(likes))
	for _, l := range likes {
		ids = append(ids, l.VideoID)
	}
	next := 0
	if len(likes) > 0 {
		next = int(likes[len(likes)-1].ID)
	}
	return ids, next, nil
}
