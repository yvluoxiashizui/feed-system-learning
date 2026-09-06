package repos

import "feed/models"

// CreateComment 创建评论
func CreateComment(c *models.Comment) error {
	return db.Create(c).Error
}

// ListCommentsByVideo 按视频查评论，最新在前
func ListCommentsByVideo(videoID uint) ([]models.Comment, error) {
	var comments []models.Comment
	if err := db.Where("video_id = ?", videoID).Order("id DESC").Find(&comments).Error; err != nil {
		return nil, err
	}
	return comments, nil
}
