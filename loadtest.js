// k6 混合场景压测脚本
// 运行：k6 run loadtest.js
// 前置：服务已启动，且数据库里有一个可登录用户和一个视频
//   默认用户 hello / 123456，默认打视频 id=1
//   可用环境变量覆盖：k6 run -e BASE_URL=... -e USER=... -e PASS=... loadtest.js
import http from 'k6/http';
import { check } from 'k6';

const BASE = __ENV.BASE_URL || 'http://localhost:8080';
const USER = __ENV.USER || 'hello';
const PASS = __ENV.PASS || '123456';
const VIDEO_ID = __ENV.VIDEO_ID || '1';

export const options = {
  vus: 50,          // 50 并发
  iterations: 5000, // 总共 5000 次请求
};

// 压测开始前先登录拿 token（发评论要用）
export function setup() {
  const res = http.post(
    `${BASE}/login`,
    JSON.stringify({ username: USER, password: PASS }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  return { token: JSON.parse(res.body).token };
}

// 混合场景：50% 刷 Feed（读缓存）+ 20% 看视频详情 + 30% 发评论（DB 写）
export default function (data) {
  const r = Math.random();
  if (r < 0.5) {
    const res = http.get(`${BASE}/videos`);
    check(res, { '200': (x) => x.status === 200 });
  } else if (r < 0.7) {
    const res = http.get(`${BASE}/video/detail?id=${VIDEO_ID}`);
    check(res, { '200': (x) => x.status === 200 });
  } else {
    const res = http.post(
      `${BASE}/comment/publish`,
      JSON.stringify({ video_id: Number(VIDEO_ID), content: 'k6 load test' }),
      { headers: { 'Content-Type': 'application/json', Authorization: data.token } }
    );
    check(res, { '200': (x) => x.status === 200 });
  }
}
