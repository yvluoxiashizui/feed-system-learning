package services

import (
	"fmt"
	"os"
	"sync"
)

// devSecret 本地开发用的兜底密钥；生产环境必须通过环境变量注入
const devSecret = "feed-dev-only-secret"

// secret 缓存解析结果：密钥在进程生命周期内是固定的，只解析一次。
// Secret 会被中间件和登录接口并发调用，用 Once 保证只读取并打印一次。
var (
	secretOnce sync.Once
	secret     []byte
)

// Secret 返回 JWT 签名密钥。
// 生产环境通过 FEED_JWT_SECRET 注入，避免密钥写死在源码里被公开仓库泄露。
// 未设置时回退到本地开发密钥，保证 clone 下来可以直接跑起来。
func Secret() []byte {
	secretOnce.Do(func() {
		if s := os.Getenv("FEED_JWT_SECRET"); s != "" {
			secret = []byte(s)
			return
		}
		fmt.Println("[warn] FEED_JWT_SECRET 未设置，正在使用本地开发密钥，请勿用于生产环境")
		secret = []byte(devSecret)
	})
	return secret
}
