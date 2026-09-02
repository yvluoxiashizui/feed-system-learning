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
	// 1. 连接 MySQL（第10课学的）
	dsn := "goapp:goapp123@tcp(127.0.0.1:3306)/feed?charset=utf8mb4&parseTime=True"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败: ", err)
	}

	// 2. 自动建表（第10课学的）
	db.AutoMigrate(&models.User{},&models.Video{})

	// 3. 把数据库连接交给 handlers 层（分层架构：main 只管组装）
	handlers.SetDB(db)

	// 4. 注册路由（第9课学的）
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

	// 预览网页：访问 http://localhost:8080/ 就是这个页面
	r.StaticFile("/", "preview.html")

	// 5. 启动服务
	r.Run(":8080")
}
