package services

import (
	"strconv"

	"github.com/redis/go-redis/v9"
)

// 令牌桶脚本。整个「取令牌 + 补令牌 + 写回」在一次 Lua 调用里完成，
// 中间不会插入其他命令；如果拆成先 GET 再 SET，两个请求可能同时读到同一个余量。
// 时间取自 Redis 的 TIME，不依赖调用方时钟，多实例部署时不会因为时钟偏差算错。
var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local cost = tonumber(ARGV[3])

local t = redis.call('TIME')
local now = tonumber(t[1]) * 1000 + math.floor(tonumber(t[2]) / 1000)

local data = redis.call('HMGET', key, 'tokens', 'ts')
local tokens = tonumber(data[1])
local ts = tonumber(data[2])

if not tokens then
  tokens = capacity
  ts = now
end

-- 按流逝时间线性补充，上限是桶容量
local elapsed = math.max(0, now - ts) / 1000
tokens = math.min(capacity, tokens + elapsed * rate)

local allowed = 0
if tokens >= cost then
  tokens = tokens - cost
  allowed = 1
end

redis.call('HSET', key, 'tokens', tokens, 'ts', now)
-- 桶从空到满所需时间的两倍，闲置后自动消失，避免键无限堆积
redis.call('PEXPIRE', key, math.ceil(capacity / rate * 1000) * 2)

return allowed
`)

// AllowToken 判断一次请求是否放行。subject 是限流对象：未登录用 IP，登录后用用户 id。
// 返回 error 表示限流组件本身不可用，由调用方决定放行还是拒绝。
func AllowToken(name, subject string, rate float64, capacity int) (bool, error) {
	key := "ratelimit:" + name + ":" + subject
	res, err := tokenBucketScript.Run(ctx, rdb, []string{key},
		strconv.FormatFloat(rate, 'f', -1, 64), capacity, 1).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}
