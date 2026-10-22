# feedsystem

基于 Go + Gin + Redis + MySQL 的短视频 Feed 流系统（简化版）。参考 [LeoninCS/feedsystem_video_go](https://github.com/LeoninCS/feedsystem_video_go) 实现核心闭环：用户注册登录、视频发布、Feed 流浏览、点赞、评论、关注、热榜。暂未包含原项目的 RabbitMQ 异步、私信、通知等模块。

## 功能

| 模块 | 功能 |
|------|------|
| 用户 | 注册、登录、JWT 签发与鉴权、按 ID 查用户 |
| 视频 | 发布视频（需登录）、Feed 流列表（分页 + Redis 缓存）、视频详情 |
| 点赞 | 点赞与取消点赞（复合唯一索引防重复）、写记录与计数同事务、是否已赞、我赞过的列表、热度累计 |
| 评论 | 发表评论、评论列表 |
| 关注 | 关注、取关、判断是否已关注、关注流（拉模式） |
| 热榜 | Redis ZSET 视频热度排行 |
| 中间件 | 请求日志、Auth JWT 鉴权 |

## 技术栈

| 分类 | 技术 |
|------|------|
| 语言 | Go |
| Web 框架 | Gin |
| 数据库 | MySQL + GORM |
| 缓存 | Redis（go-redis） |
| 认证 | JWT（golang-jwt）+ bcrypt |

## 本地开发

```bash
# 1. 建库（数据库连接配置在 main.go，本地开发使用）
mysql -u root -e "CREATE DATABASE feed CHARACTER SET utf8mb4;"
mysql -u root < mysql-setup-feed.sql   # 建库建用户（可重复执行）

# 2. 启动 Redis
service redis-server start

# 3. 启动
go run .

# 4. 预览页（刷 Feed）
# 浏览器打开 http://localhost:8080/
```

## 接口清单

### 用户
| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/register` | 否 | 注册（用户名只能字母数字，密码 bcrypt 加密） |
| POST | `/login` | 否 | 登录，返回 JWT token |
| GET | `/user?id=1` | 否 | 按 ID 查用户 |

### 视频
| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/video/publish` | JWT | 发布视频（标题 ≤100 字，播放地址 ≤500 字节） |
| GET | `/videos?limit=&cursor=` | 否 | Feed 流列表（游标分页，Redis 缓存 30s 加抖动，返回 `next_cursor`） |
| GET | `/video/detail?id=` | 否 | 视频详情 |

### 点赞
| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/video/like` | JWT | 点赞（重复返回 400），写记录与计数在同一事务，同步累计热度 |
| POST | `/video/unlike` | JWT | 取消点赞（未点赞返回 400），计数回退且不会减成负数，热度归零时移出热榜 |
| GET | `/video/isLiked?video_id=` | JWT | 是否已点赞 |
| GET | `/video/liked?limit=&cursor=` | JWT | 我赞过的视频（游标为点赞记录 id） |

### 评论
| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/comment/publish` | JWT | 发表评论（≤500 字，视频需存在） |
| GET | `/comment/list?video_id=` | 否 | 视频的评论列表 |

### 关注
| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| POST | `/social/follow` | JWT | 关注（重复返回 400） |
| POST | `/social/unfollow` | JWT | 取关（未关注返回 400） |
| GET | `/social/isFollowing?vlogger_id=` | JWT | 是否已关注 |
| GET | `/feed/following` | JWT | 关注流（关注博主的最新视频） |

### Feed / 热榜
| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/feed/hot` | 否 | 视频热榜（Redis ZSET 热度排行） |

### 测试
| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/auth-test` | JWT | 测试 Auth 中间件 |

## 性能压测

使用 [k6](https://k6.io/) 做本地混合场景压测（50 并发 / 5000 次迭代，50% Feed 列表、20% 视频详情、30% 发评论）。环境、负载模型与局限见 [docs/performance-testing.md](docs/performance-testing.md)。

| 方案 | RPS | P95 延迟 | 错误率 |
|------|-----|---------|--------|
| 无缓存（直接查 MySQL） | ~4000 | 29–30 ms | 0% |
| **Redis 缓存（Cache-Aside，30s 加抖动）** | **~5200** | **26 ms** | 0% |

> Feed 读路径本身：缓存关闭时平均 6.67 ms、P95 13.28 ms；开启后平均 2.36 ms、P95 5.97 ms。聚合 P95 由评论写路径（P95 39.67 ms）决定，读缓存对它没有作用。

```bash
# 压测命令（k6）
k6 run loadtest.js

# 对照：关闭缓存
FEED_CACHE_DISABLED=1 go run .
k6 run loadtest.js
```

## 目录结构

```
feed/
├── main.go              # 入口：连接数据库/Redis、建表、注入分层、注册路由
├── models/              # 数据模型（entity：结构体 ↔ 数据库表）
│   ├── user.go / video.go / like.go / follow.go / comment.go
├── repos/               # 数据访问层（repository）：所有 MySQL 操作
│   ├── db.go / user_repo.go / video_repo.go / like_repo.go / follow_repo.go / comment_repo.go
├── services/            # 业务逻辑层（service）：业务编排 + Redis 缓存
│   ├── redis.go / user_service.go / video_service.go / follow_service.go / comment_service.go
├── handlers/            # 接口层（handler）：收请求、调 service、返回 JSON + 中间件
│   ├── user.go / video.go / follow.go / comment.go
│   ├── auth.go          # JWT 鉴权中间件
│   └── logger.go        # 日志中间件
└── preview.html         # Feed 预览页
```

## 待办

- [x] 缓存穿透 / 击穿 / 雪崩防护
- [ ] RabbitMQ 异步
