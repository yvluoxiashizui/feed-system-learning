// k6 混合场景压测脚本
//
// 运行：
//   k6 run loadtest.js
//   k6 run -e BASE_URL=http://127.0.0.1:8080 -e VUS=50 -e ITERATIONS=5000 loadtest.js
//
// 前置：服务已启动，数据库里有一个可登录用户和一个视频。
//   默认用户 hello / 123456，默认打视频 id=1，可用环境变量覆盖。
//
// 场景配比：读 70%（Feed 首页 50% + 视频详情 20%），写 30%（发评论）。
// 每个请求带 path 标签，可以在结果里分开看读路径和写路径的延迟。
//
// 用 RANDOM_CURSOR=1 让 Feed 请求随机落在不同游标上，模拟深翻页：
//   此时缓存 key 分散、命中率下降，用于对照「首页命中」和「翻页未命中」的差异。

import http from 'k6/http';
import { check } from 'k6';

const BASE = __ENV.BASE_URL || 'http://localhost:8080';
const USER = __ENV.USER || 'hello';
const PASS = __ENV.PASS || '123456';
const VIDEO_ID = __ENV.VIDEO_ID || '1';
const VUS = Number(__ENV.VUS || 50);
const ITERATIONS = Number(__ENV.ITERATIONS || 5000);
const RANDOM_CURSOR = __ENV.RANDOM_CURSOR === '1';

export const options = {
  vus: VUS,
  iterations: ITERATIONS,
  thresholds: {
    // 错误率超过 1% 即判定本次压测不通过
    http_req_failed: ['rate<0.01'],
    checks: ['rate>0.99'],
    // 读写两条路径分别设阈值，避免写路径被读路径平均掉
    'http_req_duration{path:feed}': ['p(95)<100'],
    'http_req_duration{path:comment}': ['p(95)<100'],
  },
};

// 压测开始前登录拿 token（发评论要用），登录失败直接终止并给出原因
export function setup() {
  const res = http.post(
    `${BASE}/login`,
    JSON.stringify({ username: USER, password: PASS }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  if (res.status !== 200) {
    throw new Error(`登录失败 status=${res.status} body=${res.body}`);
  }
  const body = res.json();
  if (!body || !body.token) {
    throw new Error(`登录响应中没有 token: ${res.body}`);
  }
  return { token: body.token };
}

export default function (data) {
  const r = Math.random();

  if (r < 0.5) {
    const cursor = RANDOM_CURSOR ? Math.floor(Math.random() * 5000) : 0;
    const res = http.get(
      `${BASE}/videos?limit=10&cursor=${cursor}`,
      { tags: { path: 'feed' } }
    );
    check(res, { 'feed 200': (x) => x.status === 200 });
  } else if (r < 0.7) {
    const res = http.get(
      `${BASE}/video/detail?id=${VIDEO_ID}`,
      { tags: { path: 'detail' } }
    );
    check(res, { 'detail 200': (x) => x.status === 200 });
  } else {
    const res = http.post(
      `${BASE}/comment/publish`,
      JSON.stringify({ video_id: Number(VIDEO_ID), content: 'k6 load test' }),
      {
        headers: {
          'Content-Type': 'application/json',
          Authorization: data.token,
        },
        tags: { path: 'comment' },
      }
    );
    check(res, { 'comment 200': (x) => x.status === 200 });
  }
}
