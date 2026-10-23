package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"feed/handlers"
	"feed/models"
	"feed/repos"
	"feed/services"
)

// env 读取环境变量，未设置时用兜底值（本地开发默认配置）
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	// 数据库连接参数全部来自环境变量，未设置时回退到本地开发默认值
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True",
		env("FEED_DB_USER", "goapp"),
		env("FEED_DB_PASSWORD", "goapp123"),
		env("FEED_DB_ADDR", "127.0.0.1:3306"),
		env("FEED_DB_NAME", "feed"),
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败: ", err)
	}

	// 配置连接池，避免并发下反复建连
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取数据库连接池失败: ", err)
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 自动建表
	db.AutoMigrate(&models.User{}, &models.Video{}, &models.Like{}, &models.Follow{}, &models.Comment{})

	// 数据库连接注入 repos 数据层
	repos.SetDB(db)

	// 创建 Redis 客户端，注入 services 业务层
	rdb := redis.NewClient(&redis.Options{Addr: env("FEED_REDIS_ADDR", "localhost:6379")})
	services.SetRedis(rdb)

	// 写扩散走消息队列异步消费；连不上时退回同步扇出，服务照常启动
	if err := services.SetupFanoutMQ(); err != nil {
		log.Println("消息队列不可用，写扩散退回同步执行:", err)
	} else {
		log.Println("写扩散已切换为异步消费")
	}

	// 注册路由
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(handlers.Logger)
	r.POST("/register", handlers.Register)
	r.POST("/login", handlers.Login)
	r.GET("/healthz", handlers.Health)
	r.GET("/user", handlers.GetUser)
	r.GET("/auth-test", handlers.Auth, func(c *gin.Context) {
		userID := c.GetString("user_id")
		c.JSON(200, gin.H{"message": "已登录", "user_id": userID})
	})
	r.POST("/video/publish", handlers.Auth, handlers.PublishVideo)
	r.GET("/videos", handlers.ListVideos)
	r.GET("/video/detail", handlers.GetVideoDetail)
	r.POST("/video/like", handlers.Auth, handlers.LikeVideo)
	r.POST("/video/unlike", handlers.Auth, handlers.UnlikeVideo)
	r.GET("/video/isLiked", handlers.Auth, handlers.IsVideoLiked)
	r.GET("/video/liked", handlers.Auth, handlers.ListMyLikedVideos)
	r.POST("/social/follow", handlers.Auth, handlers.FollowUser)
	r.POST("/social/unfollow", handlers.Auth, handlers.Unfollow)
	r.GET("/social/isFollowing", handlers.Auth, handlers.IsFollowing)
	r.GET("/feed/following", handlers.Auth, handlers.FollowingFeed)
	r.GET("/feed/hot", handlers.HotVideos)
	r.POST("/comment/publish", handlers.Auth, handlers.PublishComment)
	r.GET("/comment/list", handlers.ListComments)

	// Feed 预览页
	r.StaticFile("/", "preview.html")

	// 启动服务
	if err := r.Run(env("FEED_HTTP_ADDR", ":8080")); err != nil {
		log.Fatal("启动服务失败: ", err)
	}
}
