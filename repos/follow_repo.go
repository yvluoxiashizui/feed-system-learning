package repos

import "feed/models"

// CreateFollow 创建关注关系，重复会因唯一索引报错
func CreateFollow(f *models.Follow) error {
	return db.Create(f).Error
}

// DeleteFollow 删除关注关系
func DeleteFollow(followerID, vloggerID uint) error {
	return db.Where("follower_id = ? AND vlogger_id = ?", followerID, vloggerID).
		Delete(&models.Follow{}).Error
}

// IsFollowing 判断是否已关注
func IsFollowing(followerID, vloggerID uint) (bool, error) {
	var count int64
	if err := db.Model(&models.Follow{}).
		Where("follower_id = ? AND vlogger_id = ?", followerID, vloggerID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindFollowingVloggerIDs 查出当前用户关注的所有博主 id
func FindFollowingVloggerIDs(followerID uint) ([]uint, error) {
	var follows []models.Follow
	if err := db.Where("follower_id = ?", followerID).Find(&follows).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(follows))
	for _, f := range follows {
		ids = append(ids, f.VloggerID)
	}
	return ids, nil
}
