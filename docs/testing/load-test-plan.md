# Kế hoạch kiểm thử tải & hiệu năng

## 1. Mục tiêu

1. Xác định **số request đồng thời tối đa** API có thể xử lý với p95 latency < 500ms
2. Tìm **dung lượng event buffer (Redis)** — độ sâu queue tối đa trước khi bộ nhớ hoặc xử lý bị nghẽn
3. Xác định **ngưỡng lỗi** — điểm mà tỷ lệ lỗi vượt quá 1%
4. Kiểm chứng **throughput scheduler** — xử lý batch theo kịp tốc độ nhập liệu
5. Đảm bảo **không mất dữ liệu** dưới tải liên tục

## 2. Chỉ số chính

| Chỉ số | Mục tiêu | Công cụ |
|--------|----------|---------|
| Throughput (req/s) | Thiết lập baseline, tìm max | k6 / JMeter |
| p50 latency | < 50ms | k6 |
| p95 latency | < 500ms | k6 |
| p99 latency | < 1000ms | k6 |
| Tỷ lệ lỗi | < 1% | k6 |
| Độ sâu Redis queue | Giám sát tốc độ tăng | redis-cli / Grafana |
| Bộ nhớ Redis | < 80% đã cấp phát | redis-cli INFO |
| CPU / Memory (app) | < 80% | Docker stats / Prometheus |
| Tốc độ insert DB | ≥ tốc độ drain buffer | Log ứng dụng |
| Số lượng dead letter | 0 khi tải bình thường | Endpoint queue stats |

## 3. Các kịch bản tải

### 3.1 Ma trận kịch bản

| Kịch bản | VUs | Thời lượng | Endpoint | Mục đích |
|----------|-----|------------|----------|----------|
| **S1: Baseline** | 10 | 2 phút | POST /api/v1/events | Thiết lập chỉ số baseline |
| **S2: Tải bình thường** | 50 | 5 phút | POST /api/v1/events | Mô phỏng traffic thông thường |
| **S3: Tải cao điểm** | 200 | 5 phút | POST /api/v1/events | Mô phỏng giờ cao điểm |
| **S4: Stress test** | 500 | 5 phút | POST /api/v1/events | Tìm điểm giới hạn |
| **S5: Spike test** | 10→500→10 | 3 phút | POST /api/v1/events | Đột biến traffic đột ngột |
| **S6: Batch endpoint** | 100 | 5 phút | POST /api/v1/events/batch | Throughput nhập liệu batch |
| **S7: Soak test** | 100 | 30 phút | POST /api/v1/events | Tải liên tục, kiểm tra memory leak |
| **S8: Bão hòa buffer** | 300 | 15 phút | POST /api/v1/events | Đổ đầy buffer nhanh hơn drain |

### 3.2 Mô hình tăng tải (Ramp-Up)

```
S1-S4: Tăng tuyến tính
┌─────────────────────────┐
│  VUs mục tiêu            │
│  ╱─────────────────────  │  ← Giữ trong thời lượng
│ ╱                        │
│╱                         │
├─────┬───────────────┬────┤
  30s    thời lượng     10s
 tăng     giữ         giảm

S5: Đột biến
┌─────────────────────────┐
│        ╱╲                │
│  500  ╱  ╲               │
│      ╱    ╲              │
│  10 ╱      ╲ 10          │
├────┬──┬──┬──┬────────────┤
  30s 30s 30s 30s 60s
```

## 4. Giới hạn request đồng thời

Tiến trình test để tìm trần:

| Giai đoạn | VUs | Dự kiến | Tiêu chí đạt |
|-----------|-----|---------|---------------|
| 1 | 50 | Ổn định | p95 < 200ms, 0% lỗi |
| 2 | 100 | Ổn định | p95 < 300ms, 0% lỗi |
| 3 | 200 | Gần giới hạn | p95 < 500ms, < 0.5% lỗi |
| 4 | 500 | Căng thẳng | p95 < 1000ms, < 1% lỗi |
| 5 | 1000 | Điểm gãy | Xác định kiểu lỗi |

## 5. Dung lượng Event Buffer

### 5.1 Tính toán buffer

| Tham số | Giá trị mặc định |
|---------|-------------------|
| Batch size (pop) | 1500 event |
| Chu kỳ scheduler | 90 giây |
| Tốc độ drain tối đa | ~1500 / 90s = **16.7 event/s** |
| Sub-batch insert DB | 500 dòng |
| Số lần retry tối đa | 3 |

### 5.2 Kế hoạch test bão hòa

| Test | Tốc độ nhập | Thời lượng | Tăng trưởng queue dự kiến |
|------|-------------|------------|---------------------------|
| Dưới drain | 10 req/s | 5 phút | Queue gần 0 |
| Bằng drain rate | 17 req/s | 5 phút | Queue ổn định |
| 2x drain rate | 34 req/s | 10 phút | ~10.200 backlog |
| 5x drain rate | 85 req/s | 10 phút | ~40.980 backlog |
| 10x drain rate | 170 req/s | 10 phút | ~92.000 backlog |

### 5.3 Giám sát buffer

Giám sát trong quá trình test qua `GET /api/v1/monitoring/queue-stats`:

```bash
# Poll độ sâu queue mỗi 5 giây
while true; do
  curl -s -H "x-api-key: $API_KEY" http://localhost:8080/api/v1/monitoring/queue-stats
  sleep 5
done
```

## 6. Ngưỡng lỗi

| Điều kiện | Hành động |
|-----------|-----------|
| Tỷ lệ lỗi > 1% trong 30s | Dừng test, điều tra |
| p95 > 2000ms trong 60s | Đánh dấu suy giảm hiệu năng |
| Bộ nhớ Redis > 80% | Nguy cơ OOM, dừng test |
| Dead letter queue > 0 | Điều tra lỗi xử lý |
| Số HTTP 500 > 10 | Vấn đề kết nối Redis hoặc DB |
| Độ sâu queue > 100.000 | Cảnh báo dung lượng buffer |

## 7. Thiết lập giám sát

| Tầng | Công cụ | Cần theo dõi |
|------|---------|--------------|
| Bộ tạo tải | k6 cloud / Grafana k6 | req/s, latency, lỗi |
| Ứng dụng | Prometheus + Grafana | CPU, memory, goroutine |
| Redis | redis-cli INFO, RedisInsight | Bộ nhớ, connected client, ops/s |
| PostgreSQL | pg_stat_statements | Latency query, kết nối, lock |
| Hạ tầng | Docker stats / node_exporter | CPU, memory, network I/O |
| Queue | `/api/v1/monitoring/queue-stats` | Độ sâu queue, số dead letter |

## 8. Công cụ

| Công cụ | Mục đích sử dụng |
|---------|-------------------|
| **k6** (chính) | Load test có thể viết script, kiểm tra ngưỡng |
| **JMeter** | Test dựa trên GUI, kịch bản phức tạp |
| **Grafana** | Dashboard giám sát real-time |
| **redis-cli** | Kiểm tra Redis trực tiếp |
| **psql** | Kiểm tra DB trực tiếp |
| **Docker Compose** | Môi trường test nhất quán |

## 9. Script k6 mẫu

### 9.1 Event đơn — Stress Test

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const JWT_TOKEN = __ENV.JWT_TOKEN;

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

  const params = {
    headers: {
      'Content-Type':  'application/json',
      'Authorization': `Bearer ${JWT_TOKEN}`,
    },
  };

  const res = http.post(`${BASE_URL}/api/v1/events`, payload, params);

  check(res, {
    'status là 200': (r) => r.status === 200,
    'latency < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(0.1);
}
```

### 9.2 Batch Event — Throughput Test

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const JWT_TOKEN = __ENV.JWT_TOKEN;

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
  const payload = JSON.stringify(generateBatch(100)); // batch size tối đa

  const params = {
    headers: {
      'Content-Type':  'application/json',
      'Authorization': `Bearer ${JWT_TOKEN}`,
    },
  };

  const res = http.post(`${BASE_URL}/api/v1/events/batch`, payload, params);

  check(res, {
    'status là 200': (r) => r.status === 200,
    'latency < 1s':  (r) => r.timings.duration < 1000,
  });

  sleep(0.5);
}
```

### 9.3 Test bão hòa buffer

```javascript
import http from 'k6/http';
import { check } from 'k6';

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const JWT_TOKEN = __ENV.JWT_TOKEN;
const API_KEY = __ENV.API_KEY;

export const options = {
  scenarios: {
    flood: {
      executor: 'constant-arrival-rate',
      rate: 200,              // 200 req/s — vượt xa drain rate ~17/s
      timeUnit: '1s',
      duration: '10m',
      preAllocatedVUs: 300,
      maxVUs: 500,
    },
    monitor: {
      executor: 'constant-vus',
      vus: 1,
      duration: '10m',
      exec: 'monitorQueue',
    },
  },
  thresholds: {
    'http_req_failed{scenario:flood}': ['rate<0.01'],
  },
};

export default function () {
  const payload = JSON.stringify({
    event:   'saturation_test',
    screen:  'test',
    user_id: Math.floor(Math.random() * 100000) + 1,
  });

  const res = http.post(`${BASE_URL}/api/v1/events`, payload, {
    headers: {
      'Content-Type':  'application/json',
      'Authorization': `Bearer ${JWT_TOKEN}`,
    },
  });

  check(res, { 'được chấp nhận': (r) => r.status === 200 });
}

export function monitorQueue() {
  const res = http.get(`${BASE_URL}/api/v1/monitoring/queue-stats`, {
    headers: { 'x-api-key': API_KEY },
  });

  console.log(`Thống kê queue: ${res.body}`);

  const { sleep } = require('k6');
  sleep(5);
}
```

### 9.4 Lệnh chạy

```bash
# Stress test
k6 run --env JWT_TOKEN="<token>" scripts/load-test-single.js

# Throughput batch
k6 run --env JWT_TOKEN="<token>" scripts/load-test-batch.js

# Bão hòa buffer
k6 run --env JWT_TOKEN="<token>" --env API_KEY="<key>" scripts/load-test-saturation.js

# Với báo cáo HTML
k6 run --out json=results.json scripts/load-test-single.js
```

## 10. Phân tích kết quả

### 10.1 Đọc kết quả k6

| Chỉ số | Tốt | Cảnh báo | Nghiêm trọng |
|--------|-----|----------|---------------|
| `http_req_duration p(95)` | < 200ms | 200–500ms | > 500ms |
| `http_req_duration p(99)` | < 500ms | 500ms–1s | > 1s |
| `http_req_failed` | 0% | < 0.5% | > 1% |
| `http_reqs` (throughput) | Ổn định/tăng | Chững lại | Giảm |
| `vus` so với `http_reqs` | Scale tuyến tính | Dưới tuyến tính | Phẳng/giảm |

### 10.2 Xác định điểm nghẽn

| Triệu chứng | Nguyên nhân có thể | Cách điều tra |
|--------------|---------------------|---------------|
| Latency đột biến, CPU thấp | Connection pool Redis cạn kiệt | Kiểm tra Redis `connected_clients` |
| CPU cao ở app | Overhead JSON serialization | Profile bằng pprof |
| Lỗi ở VUs cao | Giới hạn goroutine Gin/Go hoặc fd cạn kiệt | Kiểm tra `ulimit -n`, số goroutine |
| Queue tăng không kiểm soát | Drain rate < tốc độ nhập | Tăng batch size hoặc giảm interval |
| Lỗi insert DB | Connection pool bão hòa | Kiểm tra `DB_MAX_OPEN_CONNS` (mặc định: 100) |
| Có dead letter | Lỗi DB liên tục | Kiểm tra log PostgreSQL |

### 10.3 Khuyến nghị điều chỉnh

| Điểm nghẽn | Tham số | Mặc định hiện tại | Hành động đề xuất |
|-------------|---------|--------------------|--------------------|
| Drain chậm | `EVENT_BUFFER_BATCH_SIZE` | 1500 | Tăng lên 3000–5000 |
| Drain chậm | `SCHEDULER_PROCESS_INTERVAL_SECONDS` | 90 | Giảm xuống 30–60 |
| Kết nối DB | `DB_MAX_OPEN_CONNS` | 100 | Tăng nếu DB hỗ trợ |
| Bộ nhớ Redis | Redis `maxmemory` | Tùy cấu hình | Đặt 2–4GB cho buffer |
| Kết nối app | OS `ulimit -n` | 256 (macOS) | Tăng lên 65535 |
