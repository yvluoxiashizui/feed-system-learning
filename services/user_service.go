package services

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"feed/models"
	"feed/repos"
)

// Register 注册
func Register(username, password string) (*models.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}
	// 校验用户名只能包含字母和数字
	for _, ch := range username {
		isLetter := (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
		isDigit := ch >= '0' && ch <= '9'
		if !isLetter && !isDigit {
			return nil, errors.New("用户名只能包含字母和数字")
		}
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}
	u := &models.User{Username: username, Password: string(hashed)}
	if err := repos.CreateUser(u); err != nil {
		return nil, errors.New("用户名已存在")
	}
	return u, nil
}

// Login 登录，返回 token 与用户
func Login(username, password string) (string, *models.User, error) {
	u, err := repos.FindUserByName(username)
	if err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  u.ID,
		"username": u.Username,
		"exp":      time.Now().Add(72 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString([]byte(JwtSecret))
	if err != nil {
		return "", nil, errors.New("签发失败")
	}
	return tokenString, u, nil
}

// GetUser 按 ID 查用户
func GetUser(id uint) (*models.User, error) {
	return repos.FindUserByID(id)
}
