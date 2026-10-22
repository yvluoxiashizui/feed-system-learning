package services

import (
	"errors"

	"feed/models"
	"feed/repos"
)

// Follow 关注：目标用户必须存在，且不能关注自己
func Follow(followerID, vloggerID uint) error {
	if followerID == vloggerID {
		return errors.New("不能关注自己")
	}
	if _, err := repos.FindUserByID(vloggerID); err != nil {
		return errors.New("用户不存在")
	}

	f := &models.Follow{FollowerID: followerID, VloggerID: vloggerID}
	if err := repos.CreateFollow(f); err != nil {
		return errors.New("不能重复关注")
	}
	return nil
}

// Unfollow 取关：没关注过时返回错误，和取消点赞保持一致
func Unfollow(followerID, vloggerID uint) error {
	removed, err := repos.DeleteFollow(followerID, vloggerID)
	if err != nil {
		return err
	}
	if !removed {
		return errors.New("还没有关注")
	}
	return nil
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
	return repos.FindVideosByAuthorIDs(ids, 100)
}
