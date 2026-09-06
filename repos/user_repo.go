package repos

import "feed/models"

// CreateUser 创建用户
func CreateUser(u *models.User) error {
	return db.Create(u).Error
}

// FindUserByName 按用户名查用户
func FindUserByName(name string) (*models.User, error) {
	var u models.User
	if err := db.Where("username = ?", name).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// FindUserByID 按 ID 查用户
func FindUserByID(id uint) (*models.User, error) {
	var u models.User
	if err := db.Where("id = ?", id).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}
