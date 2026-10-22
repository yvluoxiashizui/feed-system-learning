package repos

import "feed/models"

// CreateComment 创建评论
func CreateComment(c *models.Comment) error {
	return db.Create(c).Error
}

// ListCommentsByVideo 按视频查评论，最新在前，游标是评论 id
func ListCommentsByVideo(videoID, cursor uint, limit int) ([]models.Comment, error) {
	var comments []models.Comment
	query := db.Where("video_id = ?", videoID).Order("id DESC").Limit(limit)
	if cursor > 0 {
		query = query.Where("id < ?", cursor)
	}
	if err := query.Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}
