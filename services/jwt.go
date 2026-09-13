package services

import (
	"fmt"
	"os"
)

// devSecret 本地开发用的兜底密钥；生产环境必须通过环境变量注入
const devSecret = "feed-dev-only-secret"

// Secret 读取 JWT 签名密钥。
// 生产环境通过 FEED_JWT_SECRET 注入，避免密钥写死在源码里被公开仓库泄露。
// 未设置时回退到本地开发密钥，保证 clone 下来可以直接跑起来。
func Secret() []byte {
	if s := os.Getenv("FEED_JWT_SECRET"); s != "" {
		return []byte(s)
	}
	fmt.Println("[warn] FEED_JWT_SECRET 未设置，正在使用本地开发密钥，请勿用于生产环境")
	return []byte(devSecret)
}
