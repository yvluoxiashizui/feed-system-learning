package services

import (
	"errors"
	"strings"
	"unicode/utf8"

	"feed/models"
	"feed/repos"
)

// maxCommentLength 单条评论的最大字符数
const maxCommentLength = 500

// PublishComment 发表评论：内容不能为空或过长，目标视频必须存在
func PublishComment(userID, videoID uint, content string) (*models.Comment, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("评论内容不能为空")
	}
	if utf8.RuneCountInString(content) > maxCommentLength {
		return nil, errors.New("评论最多 500 字")
	}
	if _, err := repos.FindVideoByID(videoID); err != nil {
		return nil, errors.New("视频不存在")
	}

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
	return repos.ListCommentsByVideo(videoID, 200)
}
