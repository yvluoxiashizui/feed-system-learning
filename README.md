# 📹 短视频 Feed 流系统（简化版）

基于 Go + Gin + GORM + MySQL 的短视频 Feed 流后端，参考 [feedsystem_video_go](https://github.com/LeoninCS/feedsystem_video_go) 做的**简化学习项目**。通过动手实现核心闭环，学习 Gin / GORM / JWT 等 Web 后端知识。

## 技术栈

| 技术 | 用途 |
|------|------|
| Go + Gin | Web 框架（路由 / handler / 中间件） |
| GORM + MySQL | 数据持久化（结构体 ↔ 数据库表） |
| JWT | 无状态认证（登录 / 鉴权中间件） |

## 已完成功能

- [x] 用户注册 / 登录（bcrypt 密码加密，用户名格式校验）
- [x] JWT 签发 + Auth 鉴权中间件（保护需要登录的接口）
- [x] 发布视频（需登录）
- [x] Feed 流（分页浏览，最新在前）

## 进行中

- [ ] 视频详情
- [ ] 点赞
- [ ] 关注

## 运行

```bash
# 1. 需要本地 MySQL，建库（开发环境配置在 main.go）
CREATE DATABASE feed CHARACTER SET utf8mb4;

# 2. 启动
go run .

# 3. 预览页（网页刷视频）
# 浏览器打开 http://localhost:8080/

# 4. 接口测试示例
# 注册
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'

# 登录拿 token
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'

# 带 token 发布视频
curl -X POST http://localhost:8080/video/publish \
  -H "Content-Type: application/json" \
  -H "Authorization: <你的token>" \
  -d '{"title":"我的视频","play_url":"/videos/1.mp4"}'

# 刷 Feed（分页）
curl "http://localhost:8080/videos?limit=10&offset=0"
```

## 目录结构

```
feed/
├── main.go              # 入口：连数据库、建表、注册路由、启动
├── models/              # 数据模型（结构体 = 数据库表）
│   ├── user.go
│   └── video.go
├── handlers/            # 业务处理（handler / 中间件）
│   ├── user.go          # 注册 / 登录 / 查用户
│   ├── video.go         # 发布视频 / Feed 列表
│   ├── logger.go        # 日志中间件
│   └── auth.go          # JWT 鉴权中间件
└── preview.html         # 预览网页
```

> 说明：数据库账号密码等配置目前为本地开发硬编码（见 main.go），生产环境应改用环境变量。
