package bootstrap

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"feed/repos"
	"feed/services"
)

// Env 读取环境变量，未设置时回退到本地开发默认值
func Env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// OpenDB 连接 MySQL、配置连接池并注入数据层。
// 迁移只在 API 进程做，Worker 只连接不建表，避免两个进程同时执行 DDL。
func OpenDB() *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True",
		Env("FEED_DB_USER", "goapp"),
		Env("FEED_DB_PASSWORD", "goapp123"),
		Env("FEED_DB_ADDR", "127.0.0.1:3306"),
		Env("FEED_DB_NAME", "feed"),
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败: ", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取数据库连接池失败: ", err)
	}
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetConnMaxLifetime(time.Hour)

	repos.SetDB(db)
	return db
}

// OpenRedis 创建 Redis 客户端并注入业务层
func OpenRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{Addr: Env("FEED_REDIS_ADDR", "localhost:6379")})
	services.SetRedis(rdb)
	return rdb
}
