package services

import (
	"errors"

	"feed/models"
	"feed/repos"
)

// Follow 关注
func Follow(followerID, vloggerID uint) error {
	f := &models.Follow{FollowerID: followerID, VloggerID: vloggerID}
	if err := repos.CreateFollow(f); err != nil {
		return errors.New("不能重复关注")
	}
	return nil
}

// Unfollow 取关
func Unfollow(followerID, vloggerID uint) error {
	return repos.DeleteFollow(followerID, vloggerID)
}

// IsFollowing 是否已关注
func IsFollowing(followerID, vloggerID uint) (bool, error) {
	return repos.IsFollowing(followerID, vloggerID)
}

// FollowingFeed 关注流：拉取关注博主的最新视频
func FollowingFeed(followerID uint) ([]models.Video, error) {
	ids, err := repos.FindFollowingVloggerIDs(followerID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []models.Video{}, nil
	}
	return repos.FindVideosByAuthorIDs(ids)
}
