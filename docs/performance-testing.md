# 压测口径

本文记录 Feed 系统 k6 压测的环境、负载模型、测量结果与局限。引用任何数字前先读「口径与局限」一节。

## 被测对象

- `GET /videos?limit=10&cursor=0`：Feed 首页，读 Redis 缓存
- `GET /video/detail?id=1`：视频详情，直接查 MySQL
- `POST /comment/publish`：发表评论，写 MySQL

## 环境

- WSL2 Ubuntu，内核 6.18.33.2-microsoft-standard-WSL2
- CPU：Intel Core i9-14900HX，WSL 可见 32 个逻辑核
- 内存：7 GB
- Go 1.26.5，MySQL 8.4.11，Redis 8.0.5，k6 v2.2.0
- 服务、MySQL、Redis 运行在同一台机器，走 loopback，无跨机网络开销

## 数据集

- videos 5002 条，users 3 条
- comments 从 0 开始，每次压测约新增 1500 条，压测前应清理

## 负载模型

- 50 个虚拟用户，固定 5000 次迭代
- 50% 请求 Feed 首页，20% 请求视频详情，30% 发表评论
- 每次运行的 RPS = 5000 / 总耗时

## 复现

```bash
# 启动服务
FEED_HTTP_ADDR=:8080 go run .

# 压测
k6 run loadtest.js
k6 run -e BASE_URL=http://127.0.0.1:8080 -e VUS=50 -e ITERATIONS=5000 loadtest.js
```

对照实验通过环境变量切换，两次运行使用同一个二进制：

```bash
FEED_CACHE_DISABLED=1   # 关闭 Feed 缓存读写
FEED_ACCESS_LOG=0       # 关闭每请求访问日志
```

## 结果

默认配置（访问日志开启、缓存开启）：

- 聚合：RPS 约 5200，P95 约 26.4 ms，错误率 0%
- Feed 读路径：平均 2.36 ms，P95 5.97 ms
- 评论写路径：平均 22.28 ms，P95 39.67 ms

关闭 Feed 缓存（`FEED_CACHE_DISABLED=1`）：

- 聚合：RPS 约 4000，P95 约 29–30.6 ms，错误率 0%
- Feed 读路径：平均 6.67 ms，P95 13.28 ms
- 评论写路径：平均 23.37 ms，P95 46.12 ms

关闭访问日志后聚合 RPS 提升约 2%，在运行波动范围内。

## 结论

缓存改善的是读路径：Feed 首页平均延迟下降约 65%，P95 下降约 55%，聚合 RPS 提升约 30%。

聚合 P95 由评论写路径决定，评论写入 P95 为 39.67 ms，占请求量的 30%。读缓存对这条路径没有作用，改善它需要从写入侧处理。

## 口径与局限

- 服务、MySQL、Redis 同机运行，互相争抢 CPU 与内存，数字不代表生产性能
- 单进程单实例，没有负载均衡和连接池调优
- 同一配置连续两轮之间波动约 10%，引用时应取整到百位，不要使用四位有效数字
- 固定迭代数，结果对总耗时敏感；对比两组时应保证迭代数和并发数一致
- 客户端与服务端在同一台机器，没有真实网络的延迟与丢包
- WSL2 网络栈有额外开销，与裸机 Linux 会有差异
- Feed 请求固定 `cursor=0`，命中同一条缓存；用 `RANDOM_CURSOR=1` 可模拟深翻页，此时缓存 key 分散、命中率下降

## 清理测试数据

```sql
DELETE FROM comments WHERE content = 'k6 load test';
DELETE FROM videos WHERE title LIKE 'seed-%';
```
