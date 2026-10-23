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

	// 关注时回填对方最近的视频，否则只能看到关注之后发布的内容
	backfillTimeline(followerID, vloggerID)
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

	// 取关时把对方的视频从时间线里摘掉，避免读到已取关的人
	clearAuthorFromTimeline(followerID, vloggerID)
	return nil
}

// IsFollowing 是否已关注
func IsFollowing(followerID, vloggerID uint) (bool, error) {
	return repos.IsFollowing(followerID, vloggerID)
}

// FollowingFeed 关注流：推来的内容读时间线，大 V 的内容实时聚合，两路归并后分页
func FollowingFeed(followerID, cursor uint, limit int) (feedPage, error) {
	following, err := repos.FindFollowingVloggerIDs(followerID)
	if err != nil {
		return feedPage{}, err
	}
	if len(following) == 0 {
		return feedPage{Videos: []models.Video{}}, nil
	}

	// 缓存关闭时没有时间线可用，退回纯拉取
	if !cacheEnabled() {
		return pullFollowingFeed(following, cursor, limit)
	}

	// 按粉丝数把关注的作者分两组：没超过阈值的已经写进时间线，超过的要实时拉
	counts, err := repos.CountFollowersByIDs(following)
	if err != nil {
		return feedPage{}, err
	}
	pullIDs := make([]uint, 0, len(following))
	for _, id := range following {
		if int(counts[id]) > fanoutThreshold() {
			pullIDs = append(pullIDs, id)
		}
	}

	// 第一路：时间线里被推来的内容
	var pushed []models.Video
	if ids, err := timelinePage(followerID, cursor, limit); err == nil && len(ids) > 0 {
		pushed, _ = loadVideosByIDs(ids)
	}

	// 第二路：大 V 的内容实时查
	var pulled []models.Video
	if len(pullIDs) > 0 {
		pulled, err = repos.FindVideosByAuthorIDs(pullIDs, cursor, limit)
		if err != nil {
			return feedPage{}, err
		}
	}

	videos := mergeByIDDesc(pushed, pulled, limit)
	page := feedPage{Videos: videos}
	if len(videos) > 0 {
		page.NextCursor = int(videos[len(videos)-1].ID)
	}
	return page, nil
}

// pullFollowingFeed 纯拉取：按关注列表实时查视频，游标是视频 id
func pullFollowingFeed(authorIDs []uint, cursor uint, limit int) (feedPage, error) {
	videos, err := repos.FindVideosByAuthorIDs(authorIDs, cursor, limit)
	if err != nil {
		return feedPage{}, err
	}
	page := feedPage{Videos: videos}
	if len(videos) > 0 {
		page.NextCursor = int(videos[len(videos)-1].ID)
	}
	return page, nil
}

// mergeByIDDesc 归并两路按 id 倒序的视频，取前 limit 条；同一个 id 只保留一个，
// 防止阈值变化后同一视频既在时间线里又被实时拉到
func mergeByIDDesc(a, b []models.Video, limit int) []models.Video {
	out := make([]models.Video, 0, limit)
	i, j := 0, 0
	for len(out) < limit && (i < len(a) || j < len(b)) {
		switch {
		case j >= len(b) || (i < len(a) && a[i].ID > b[j].ID):
			out = append(out, a[i])
			i++
		case i >= len(a) || b[j].ID > a[i].ID:
			out = append(out, b[j])
			j++
		default:
			out = append(out, a[i])
			i++
			j++
		}
	}
	return out
}
