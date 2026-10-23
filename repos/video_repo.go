package repos

import (
	"gorm.io/gorm"

	"feed/models"
)

// CreateVideo 创建视频
func CreateVideo(v *models.Video) error {
	return db.Create(v).Error
}

// ListVideosByCursor 游标分页查视频，最新在前
func ListVideosByCursor(cursor uint, limit int) ([]models.Video, error) {
	var videos []models.Video
	query := db.Order("id DESC").Limit(limit)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	if err := query.Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

// ListRecentVideoIDs 取最近 N 条视频的 id，按 id 降序
func ListRecentVideoIDs(limit int) ([]uint, error) {
	ids := make([]uint, 0, limit)
	if err := db.Model(&models.Video{}).
		Order("id DESC").
		Limit(limit).
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// FindVideoByID 按 ID 查视频
func FindVideoByID(id uint) (*models.Video, error) {
	var v models.Video
	if err := db.Where("id = ?", id).First(&v).Error; err != nil {
		return nil, err
	}
	return &v, nil
}

// IncrementVideoLikes 视频点赞数 +1（原子自增）
func IncrementVideoLikes(videoID uint) error {
	return db.Model(&models.Video{}).
		Where("id = ?", videoID).
		UpdateColumn("likes_count", gorm.Expr("likes_count + 1")).Error
}

// FindVideosByIDs 按 id 列表查视频
func FindVideosByIDs(ids []uint) ([]models.Video, error) {
	var videos []models.Video
	if err := db.Where("id IN ?", ids).Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

// FindRecentVideoIDsByAuthor 某作者最近的 N 条视频 id，降序
func FindRecentVideoIDsByAuthor(authorID uint, limit int) ([]uint, error) {
	ids := make([]uint, 0, limit)
	if err := db.Model(&models.Video{}).
		Where("author_id = ?", authorID).
		Order("id DESC").
		Limit(limit).
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// FindVideoIDsByAuthor 某作者的全部视频 id
func FindVideoIDsByAuthor(authorID uint) ([]uint, error) {
	ids := make([]uint, 0, 64)
	if err := db.Model(&models.Video{}).
		Where("author_id = ?", authorID).
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

// FindVideosByAuthorIDs 按作者 id 列表查视频，最新在前，游标是视频 id
func FindVideosByAuthorIDs(authorIDs []uint, cursor uint, limit int) ([]models.Video, error) {
	var videos []models.Video
	query := db.Where("author_id IN ?", authorIDs).Order("id DESC").Limit(limit)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	if err := query.Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}
