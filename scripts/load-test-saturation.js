import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  scenarios: {
    flood: {
      executor: 'constant-arrival-rate',
      rate: 200,
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 300,
      maxVUs: 500,
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

export default function () {
  const payload = JSON.stringify({
    event:   'saturation_test',
    screen:  'test',
    user_id: Math.floor(Math.random() * 100000) + 1,
    occurred_at: new Date().toISOString(),
  });

  const res = http.post(`${BASE_URL}/api/v1/events`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, { 'status 200': (r) => r.status === 200 });
}
