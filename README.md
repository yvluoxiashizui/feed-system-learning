# 短视频 Feed 流系统（简化版）

一个学习后端的练手项目。参考了 [LeoninCS/feedsystem_video_go](https://github.com/LeoninCS/feedsystem_video_go) 做了个简化版，想练练 Gin、GORM、MySQL、JWT 这些。

现在还在慢慢做，先把核心流程走通。

## 目前能做的

- 注册 / 登录（密码是 bcrypt 加密后存的，不会存明文）
- 登录之后才能发视频（用 JWT 验证）
- 刷视频列表，带分页

## 用的东西

- Go + Gin：处理 HTTP 请求，路由、中间件
- GORM + MySQL：结构体对应数据库表，增删改查
- JWT：登录后发个 token，之后请求带上它验证身份

## 怎么跑起来

1. 本地装好 MySQL，建一个库：
```sql
CREATE DATABASE feed CHARACTER SET utf8mb4;
```
2. 数据库连接配置在 main.go 里（开发用的）
3. 跑起来：
```bash
go run .
```
4. 浏览器打开 http://localhost:8080/ 能看到一个简单预览页

## 目录

```
feed/
├── main.go          # 入口：连库、建表、路由
├── models/          # 数据模型（User、Video）
├── handlers/        # 接口逻辑（注册登录、发视频、中间件）
└── preview.html     # 简单预览页
```

## 还在做 / 想做的

- [ ] 视频详情
- [ ] 点赞
- [ ] 关注

## 踩过的坑（自己记一下）

- GORM 的 `db.Create()` 后面要加 `.Error` 才能拿到错误，不然会"存进去了还报错"
- 路由注册要写在 main 顶层，不能写进 handler 函数里面
- json 标签要用小写（`json:"title"`），写成大写 `JSON` 会绑定不上
- 新加的模型要记得加进 `AutoMigrate`，不然表不会建

## 说明

数据库账号密码目前是写死在代码里的开发配置，以后做成真实项目会用环境变量。
