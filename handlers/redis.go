package handlers

import (
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client
func SetRedis(r *redis.Client) {
	rdb = r
}