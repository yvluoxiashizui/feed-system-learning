package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"feed/handlers"
	"feed/models"
)

func main() {
	// 连接数据库（开发配置，生产环境建议改用环境变量）
	dsn := "goapp:goapp123@tcp(127.0.0.1:3306)/feed?charset=utf8mb4&parseTime=True"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败: ", err)
	}

	// 自动建表
	db.AutoMigrate(&models.User{}, &models.Video{}, &models.Like{})

	// 数据库连接注入 handlers 层
	handlers.SetDB(db)

	// 注册路由
	r := gin.Default()
	r.Use(handlers.Logger)
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)
	r.GET("/user", handlers.GetUser)
	r.GET("/auth-test", handlers.Auth, func(c *gin.Context) {
		userID := c.GetString("user_id")
		c.JSON(200, gin.H{"message": "已登录", "user_id": userID})
	})
	r.POST("/video/publish", handlers.Auth, handlers.PublishVideo)
	r.GET("/videos", handlers.ListVideos)
	r.GET("/video/detail", handlers.GetVideoDetail)
	r.POST("/video/like", handlers.Auth, handlers.LikeVideo)

	// Feed 预览页
	r.StaticFile("/", "preview.html")

	// 启动服务
	r.Run(":8080")
}
