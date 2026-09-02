# feedsystem

基于 Go + Gin 的短视频 Feed 流系统（简化版）。参考 [LeoninCS/feedsystem_video_go](https://github.com/LeoninCS/feedsystem_video_go) 实现核心闭环：用户注册登录、视频发布、Feed 流浏览。简化版暂未包含原项目的 Redis 缓存、RabbitMQ 异步、私信、通知等模块。

## 功能

| 模块 | 功能 |
|------|------|
| 用户 | 注册、登录、JWT 签发与鉴权 |
| 视频 | 发布视频（需登录）、Feed 流列表（分页） |
| 中间件 | 请求日志、Auth 鉴权 |

## 技术栈

| 分类 | 技术 |
|------|------|
| 语言 | Go |
| Web 框架 | Gin |
| 数据库 | MySQL + GORM |
| 认证 | JWT（golang-jwt） |

## 本地开发

```bash
# 1. 建库（数据库连接配置在 main.go，本地开发使用）
mysql -u root -e "CREATE DATABASE feed CHARACTER SET utf8mb4;"

# 2. 启动
go run .

# 3. 预览页（刷 Feed）
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
| POST | `/video/publish` | JWT | 发布视频 |
| GET | `/videos?limit=&offset=` | 否 | Feed 流列表（分页，最新在前） |

### 测试
| 方法 | 路径 | 鉴权 | 说明 |
|------|------|------|------|
| GET | `/auth-test` | JWT | 测试 Auth 中间件 |

## 目录结构

```
feed/
├── main.go              # 入口：连接数据库、建表、注册路由
├── models/              # 数据模型（结构体 ↔ 数据库表）
│   ├── user.go          # User 用户表
│   └── video.go         # Video 视频表
├── handlers/            # 业务处理
│   ├── user.go          # 注册 / 登录 / 查用户
│   ├── video.go         # 发布视频 / Feed 列表
│   ├── auth.go          # JWT 鉴权中间件
│   └── logger.go        # 日志中间件
└── preview.html         # Feed 预览页
```

## 待办

- [ ] 视频详情
- [ ] 点赞
- [ ] 关注
