import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export const options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '3m',  target: 100 },
    { duration: '30s', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'],
    http_req_failed:   ['rate<0.01'],
  },
};

function generateBatch(size) {
  const events = [];
  for (let i = 0; i < size; i++) {
    events.push({
      event:   'batch_load_test',
      screen:  'test_screen',
      user_id: Math.floor(Math.random() * 100000) + 1,
      properties: { index: i },
      occurred_at: new Date().toISOString(),
    });
  }
  return { events };
}

export default function () {
  const payload = JSON.stringify(generateBatch(100));

  const res = http.post(`${BASE_URL}/api/v1/events/batch`, payload, {
    headers: { 'Content-Type': 'application/json' },
  });

  check(res, {
    'status 200': (r) => r.status === 200,
    'latency < 1s': (r) => r.timings.duration < 1000,
  });

  sleep(0.5);
}
