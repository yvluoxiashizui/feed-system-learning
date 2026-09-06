package services

import (
	"errors"

	"feed/models"
	"feed/repos"
)

// PublishComment 发表评论
func PublishComment(userID, videoID uint, content string) (*models.Comment, error) {
	author, err := repos.FindUserByID(userID)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	c := &models.Comment{UserID: userID, Username: author.Username, VideoID: videoID, Content: content}
	if err := repos.CreateComment(c); err != nil {
		return nil, err
	}
	return c, nil
}

// ListComments 评论列表
func ListComments(videoID uint) ([]models.Comment, error) {
	return repos.ListCommentsByVideo(videoID)
}
