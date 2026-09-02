package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"feed/models"
)

// 全局数据库连接，由 main.go 传入
var db *gorm.DB

// SetDB 把 main 里的数据库连接交给 handlers 用
func SetDB(d *gorm.DB) {
	db = d
}

// Register 注册：接收用户名密码 → 密码加密 → 存数据库
func Register(c *gin.Context) {
	// 1. 把请求体里的 JSON 解析到 input（你学过的 ShouldBindJSON）
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	
	if input.Username == "" || input.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名和密码不能为空"})
		return
	}

	for _,ch := range input.Username{
		ifabc := (ch >= 'a'&&ch <= 'z')||(ch >= 'A'&&ch <= 'Z')
		ifnum := ch>='0'&&ch<='9' 
		if !ifabc && !ifnum{
			c.JSON(http.StatusBadRequest,gin.H{"error":"用户名只能包含数字或字母"})
		return
		}
	}

	// 2. 密码加密（bcrypt）——【新知识点】数据库里绝不存明文密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "加密失败"})
		return
	}

	// 3. 存进数据库
	user := models.User{Username: input.Username, Password: string(hashed)}
	if err := db.Create(&user).Error; err != nil {
		// 用户名重复会触发 uniqueIndex 报错
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}

	// 4. 返回（注意：不会返回 password，因为字段上写了 json:"-"）
	c.JSON(http.StatusOK, gin.H{"id": user.ID, "username": user.Username})
}

// Login 登录：查用户 → 比对密码 → 签发 JWT token
func Login(c *gin.Context) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	// 1. 按用户名查用户（你学过的 Where + First）
	var user models.User
	if err := db.Where("username = ?", input.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 2. 比对密码（bcrypt 校验）
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 3. 【新知识点】签发 JWT token——一张"带签名的身份卡"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,                                  // 里面装谁
		"username": user.Username,
		"exp":      time.Now().Add(72 * time.Hour).Unix(), // 72 小时后过期
	})
	tokenString, err := token.SignedString([]byte("feed-secret-key")) // 用密钥签名
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "签发失败"})
		return
	}

	// 4. 把 token 返回给前端，以后访问带这个 token 就行
	c.JSON(http.StatusOK, gin.H{"token": tokenString, "user_id": user.ID, "username": user.Username})
}

// GetUser 按 ID 查用户（演示 c.Query 取参数 + c.JSON 返回）
func GetUser(c *gin.Context) {
	id := c.Query("id") // 从 URL 里取 ?id=xxx

	var user models.User
	// 按 id 查（你学过的 Where + First）
	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	// 返回用户（密码不会出现，因为字段上写了 json:"-"）
	c.JSON(http.StatusOK, user)
}
