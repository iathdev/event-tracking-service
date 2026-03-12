import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '2m',  target: 200 },
    { duration: '2m',  target: 500 },
    { duration: '1m',  target: 500 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed:   ['rate<0.01'],
  },
};

export default function () {
  const payload = JSON.stringify({
    event:   'load_test_event',
    screen:  'test_screen',
    user_id: Math.floor(Math.random() * 100000) + 1,
    properties: { test_run: `${__ENV.K6_TEST_RUN || 'default'}` },
    occurred_at: new Date().toISOString(),
  });

  const res = http.post(`${BASE_URL}/api/v1/events`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, {
    'status 200': (r) => r.status === 200,
    'latency < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(0.1);
}
