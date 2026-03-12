# Báo cáo kết quả Load Test

**Ngày chạy:** 2026-03-15
**Môi trường:** Local (macOS, localhost:8080)
**Công cụ:** k6 v1.6.1
**Service:** Event Tracking Service (Go/Gin + Redis buffer)

---

## Đánh giá cấu hình hiện tại

**Cấu hình scheduler:**

| Tham số | Giá trị |
|---------|---------|
| `SCHEDULER_PROCESS_INTERVAL_SECONDS` | 60 |
| `SCHEDULER_PROCESS_TIMEOUT_SECONDS` | 60 |
| `EVENT_BUFFER_BATCH_SIZE` | 1500 |
| DB sub-batch (GORM `CreateInBatches`) | 1000 rows |
| `EVENT_BUFFER_MAX_RETRIES` | 3 |
| Redis TTL cho event queue | **Không có TTL** (tồn tại vĩnh viễn cho đến khi được pop) |

**Tốc độ drain lý thuyết:** 1500 event / 60s = **25 event/s**

### Khả năng đáp ứng theo mức traffic

| Mức traffic | Tốc độ nhập | So với drain (25/s) | Backlog sau 1 giờ | Đánh giá |
|-------------|-------------|---------------------|---------------------|----------|
| Rất thấp | 5 req/s | 0.2x | 0 (drain kịp) | Thoải mái |
| Thấp | 15 req/s | 0.6x | 0 (drain kịp) | Thoải mái |
| Vừa đủ | 25 req/s | 1x | 0 (cân bằng) | Giới hạn |
| Trung bình | 50 req/s | 2x | ~90.000 event | Tồn đọng, tự drain khi traffic giảm |
| Cao (peak test) | 200 req/s | 8x | ~630.000 event | Tồn đọng, tự drain khi traffic giảm |
| Stress (test) | 2.450 req/s | 98x | ~8.7 triệu event | Tồn đọng, tự drain khi traffic giảm |

### Khi traffic vượt drain rate — có mất dữ liệu không?

**Không.** Hệ thống được thiết kế async theo mô hình event buffer:

1. API chỉ push event vào Redis List (`RPush`) rồi trả **202 Accepted** ngay — không phụ thuộc tốc độ drain.
2. Event trong Redis **không có TTL**, tồn tại vĩnh viễn cho đến khi scheduler pop ra xử lý.
3. Khi traffic giảm (ví dụ hết giờ thi), scheduler sẽ **tự động drain hết backlog** theo từng batch 1500 event / 60s.
4. Nếu insert DB thất bại sau 3 lần retry → event được chuyển sang **dead letter queue**, không bị mất.

**Ví dụ:** Peak 200 req/s trong 10 phút → tích lũy ~120.000 event → traffic về 0 → scheduler drain hết trong ~80 phút (120.000 / 25 event/s).

### Khi Redis die — ảnh hưởng như thế nào?

| Luồng | Ảnh hưởng | Chi tiết |
|-------|-----------|----------|
| **API (nhận event)** | `RPush` thất bại → trả **500** | Client không gửi được event mới. App không crash, `/health` và các endpoint khác vẫn hoạt động |
| **Scheduler (xử lý event)** | Không acquire được Redis lock → **job bị skip** | Event đã có trong queue chưa kịp pop, phụ thuộc Redis persistence |
| **Monitoring** | `LLen` thất bại → `queue_size: -1` | Endpoint vẫn trả 200 nhưng giá trị = -1 |

**Sau khi Redis phục hồi:**
- Nếu bật `appendonly yes` (AOF) → queue còn nguyên, scheduler tự chạy lại, **không mất event**.
- Nếu không bật persistence → **mất toàn bộ queue** chưa xử lý.

**Rủi ro và khuyến nghị:**

| Rủi ro | Hậu quả | Khuyến nghị |
|--------|---------|-------------|
| Redis restart không có persistence | Mất toàn bộ event trong queue | Bật `appendonly yes` trên production |
| Redis hết memory (`maxmemory`) | API trả 500, không nhận event mới | Set `maxmemory` đủ lớn, giám sát usage |
| Redis die kéo dài | Mất event trong khoảng thời gian die (client nhận 500) | Thiết lập Redis Sentinel hoặc Redis Cluster cho HA |

> **Tóm lại:** Redis die = mất khả năng nhận event mới + dừng xử lý event cũ. Nhưng app không crash, không ảnh hưởng DB. Redis lên lại là tự phục hồi.

### Phương án fallback khi Redis die (chưa implement)

**Hiện trạng:** Khi Redis die, handler trả 500 và **bỏ qua event hoàn toàn** — không có cơ chế fallback.

**Đề xuất: Fallback ghi thẳng DB có rate limit**

```
Luồng bình thường:
  HTTP → Redis RPush → 202 Accepted

Khi Redis die (RPush fail):
  HTTP → In-memory batch buffer (gom tối đa 500 event hoặc 5s)
       → INSERT trực tiếp PostgreSQL (rate limit ~100 event/s)
       → Nếu DB cũng fail → trả 503 Service Unavailable

Khi Redis phục hồi:
  → Tự động chuyển lại luồng bình thường
```

**So sánh các phương án fallback:**

| Phương án | Giữ event | Bảo vệ DB | Độ phức tạp | Mất event khi app restart |
|-----------|-----------|-----------|-------------|---------------------------|
| In-memory buffer | Giữ trong RAM, flush batch vào DB | Rate limit được | Thấp | Có |
| Local file buffer | Ghi ra file, replay khi Redis lên | Không đụng DB | Trung bình | Không |
| **Direct DB có rate limit (đề xuất)** | **Ghi thẳng DB, giới hạn tốc độ** | **Có giới hạn** | **Trung bình** | **Không** |
| Kết hợp RAM + file | RAM trước, tràn thì ghi file | Rate limit + không mất | Cao | Không |

**Lý do chọn Direct DB có rate limit:**
- Không thêm dependency mới (file system, message queue)
- Event đến đúng đích cuối (PostgreSQL) — không mất khi app restart
- Gom 500 event/batch = 1 DB connection thay vì 500 → DB chịu được
- Rate limit ~100 event/s đủ cho hầu hết traffic thực tế
- `DB_MAX_OPEN_CONNS` = 100, chỉ dùng 1 connection cho fallback → không ảnh hưởng query khác

### Kịch bản thực tế

Với hệ thống B2B Exam Platform (IELTS/TOEIC), traffic tracking event phụ thuộc vào số lượng thí sinh online đồng thời:

| Số thí sinh đồng thời | Event ước tính (1 event/3s) | Cấu hình hiện tại đáp ứng? |
|------------------------|----------------------------|----------------------------|
| 50 | ~17 req/s | Đáp ứng tốt |
| 75 | ~25 req/s | Vừa đủ (giới hạn) |
| 100+ | ~33+ req/s | Bắt đầu tồn đọng |

### Khuyến nghị nâng cấp khi cần

| Khi nào | Hành động | Drain rate mới |
|---------|-----------|----------------|
| > 75 thí sinh đồng thời | Tăng `EVENT_BUFFER_BATCH_SIZE` → 3000 | ~50/s |
| > 150 thí sinh đồng thời | Tăng `EVENT_BUFFER_BATCH_SIZE` → 5000 | ~83/s |
| > 300 thí sinh đồng thời | Tăng batch → 5000 + giảm interval → 30s | ~167/s |

> **Kết luận:** Cấu hình hiện tại (interval=60s, batch=1500) phù hợp cho **tối đa ~75 thí sinh đồng thời**. Khi scale lên, ưu tiên tăng `EVENT_BUFFER_BATCH_SIZE` trước, giảm interval sau.

---

## Tổng quan

| # | Kịch bản | Kết quả | Tổng request | Lỗi |
|---|----------|---------|--------------|------|
| 1 | Stress test — Event đơn (50→500 VUs) | PASS | 881.805 | 0% |
| 2 | Throughput — Batch 100 event/req (50→100 VUs) | PASS | 25.676 | 0% |
| 3 | Bão hòa buffer — 200 req/s cố định | PASS | 60.001 | 0% |

---

## Test 1: Stress Test — Event đơn

**Script:** `scripts/load-test-single.js`
**Endpoint:** `POST /api/v1/events`
**Mô tả:** Tăng tải từ 50 → 200 → 500 VUs, giữ 500 VUs trong 1 phút, sau đó giảm về 0. Tổng ~6 phút.

### Kết quả

| Chỉ số | Giá trị | Ngưỡng | Đánh giá |
|--------|---------|--------|----------|
| Tổng request | 881.805 | — | — |
| Throughput | **2.449 req/s** | — | Rất cao |
| Tỷ lệ lỗi HTTP | **0.00%** | < 1% | PASS |
| p50 (median) | **2.55ms** | — | Xuất sắc |
| p90 | **6.66ms** | — | Xuất sắc |
| p95 | **8.81ms** | < 500ms | PASS |
| p99 | **205.39ms** | < 1000ms | PASS |
| Max | 1.4s | — | Chấp nhận được (outlier) |
| Data gửi | 244 MB | — | — |
| Data nhận | 138 MB | — | — |

### Check assertions

| Check | Kết quả | Chi tiết |
|-------|---------|----------|
| Status 202 | 100% | 881.805 / 881.805 |
| Latency < 500ms | 99.94% | 881.275 / 881.805 (530 request > 500ms) |

### Nhận xét

- API xử lý **500 VUs đồng thời** mà không có lỗi HTTP nào.
- p95 chỉ **8.81ms** — cực kỳ nhanh, do API chỉ push vào Redis rồi trả 202 ngay.
- 530 request có latency > 500ms (0.06%) — đây là outlier, max 1.4s, có thể do GC hoặc Redis connection spike.
- Throughput đạt **~2.450 req/s** ổn định.

---

## Test 2: Throughput — Batch Event

**Script:** `scripts/load-test-batch.js`
**Endpoint:** `POST /api/v1/events/batch` (100 event/request)
**Mô tả:** Tăng từ 50 → 100 VUs, giữ 100 VUs trong 3 phút. Mỗi request gửi 100 event. Tổng ~4 phút.

### Kết quả

| Chỉ số | Giá trị | Ngưỡng | Đánh giá |
|--------|---------|--------|----------|
| Tổng request | 25.676 | — | — |
| Tổng event (ước tính) | **~2.567.600** | — | — |
| Throughput | **106.86 req/s** (~10.686 event/s) | — | Cao |
| Tỷ lệ lỗi HTTP | **0.00%** | < 1% | PASS |
| p50 (median) | **7.2ms** | — | Tốt |
| p90 | **19.41ms** | — | Tốt |
| p95 | **203.6ms** | < 1000ms | PASS |
| Max | 22.89s | — | Cần lưu ý |
| Data gửi | 340 MB | — | — |
| Data nhận | 4.2 MB | — | — |

### Check assertions

| Check | Kết quả | Chi tiết |
|-------|---------|----------|
| Status 202 | 100% | 25.676 / 25.676 |
| Latency < 1s | 98.06% | 25.177 / 25.676 (499 request > 1s) |

### Nhận xét

- **Không có lỗi HTTP** — tất cả 25.676 batch request đều trả 202.
- Throughput batch: **~107 req/s × 100 event = ~10.700 event/s** đẩy vào Redis.
- p95 = 203.6ms — vẫn tốt, nhưng max lên tới **22.89s** cho thấy có một số request bị latency spike.
- 499 request (1.94%) vượt ngưỡng 1s — có thể do Redis pipeline bị chậm khi push 100 event cùng lúc ở tải cao.
- So với event đơn: throughput event/s cao hơn gấp **~4.4 lần** (10.700 vs 2.450), batch hiệu quả hơn rõ rệt.

---

## Test 3: Bão hòa Buffer

**Script:** `scripts/load-test-saturation.js`
**Endpoint:** `POST /api/v1/events` với tốc độ cố định 200 req/s
**Mô tả:** Gửi đều 200 request/s trong 5 phút để kiểm tra khả năng chịu tải ổn định và tích lũy buffer.

### Kết quả

| Chỉ số | Giá trị | Ngưỡng | Đánh giá |
|--------|---------|--------|----------|
| Tổng request | 60.001 | — | — |
| Throughput thực tế | **200.00 req/s** | Mục tiêu 200 | Đạt chính xác |
| Tỷ lệ lỗi HTTP | **0.00%** | < 1% | PASS |
| p50 (median) | **734µs** | — | Xuất sắc |
| p90 | **1.1ms** | — | Xuất sắc |
| p95 | **1.35ms** | < 500ms | PASS |
| Max | 30.74ms | — | Xuất sắc |
| VUs sử dụng thực tế | **0–1** | Cấp phát 300 | Rất nhàn |
| Data gửi | 14 MB | — | — |
| Data nhận | 9.4 MB | — | — |

### Check assertions

| Check | Kết quả | Chi tiết |
|-------|---------|----------|
| Status 202 | **100%** | 60.001 / 60.001 |

### Phân tích buffer

| Tham số | Giá trị |
|---------|---------|
| Tốc độ nhập | 200 event/s |
| Tốc độ drain (lý thuyết) | ~16.7 event/s (1500 event / 90s) |
| Tổng event nhập trong 5 phút | 60.001 |
| Tổng event drain trong 5 phút (lý thuyết) | ~5.000 (3 chu kỳ × 1500 + phần lẻ) |
| **Backlog ước tính sau 5 phút** | **~55.000 event trong Redis queue** |

### Nhận xét

- API xử lý 200 req/s **hoàn toàn nhẹ nhàng** — chỉ cần 0–1 VU (k6 tự tối ưu vì response quá nhanh).
- p95 chỉ **1.35ms** — nhanh nhất trong 3 test vì tải phân bổ đều, không có spike.
- **Vấn đề chính:** Tốc độ nhập (200/s) >> tốc độ drain (~17/s), sau 5 phút ước tính **~55.000 event tồn đọng** trong Redis.
- Nếu chạy lâu hơn, queue sẽ tiếp tục tăng → cần tăng `EVENT_BUFFER_BATCH_SIZE` hoặc giảm `SCHEDULER_PROCESS_INTERVAL_SECONDS`.

---

## Tổng kết

### Điểm mạnh

- **0% lỗi HTTP** trên cả 3 kịch bản — API rất ổn định.
- **Latency cực thấp** cho endpoint event đơn: p95 < 10ms ở 500 VUs.
- Kiến trúc async (HTTP → Redis → DB) giúp API luôn phản hồi nhanh dù DB chậm.
- Batch endpoint hiệu quả gấp **~4.4 lần** so với gửi từng event.

### Điểm cần cải thiện

| Vấn đề | Mức độ | Khuyến nghị |
|--------|--------|-------------|
| Batch max latency 22.89s | Trung bình | Kiểm tra Redis pipeline performance khi push 100 event |
| Buffer drain chậm hơn ingestion | Cao | Tăng `EVENT_BUFFER_BATCH_SIZE` (1500 → 3000–5000) hoặc giảm `SCHEDULER_PROCESS_INTERVAL_SECONDS` (90 → 30–60) |
| Chưa test với Redis/DB thực tế remote | — | Cần test thêm trên staging với network latency thật |

### Năng lực hệ thống (local)

| Chỉ số | Giá trị đo được |
|--------|------------------|
| Max concurrent VUs (0% error) | **500+** |
| Max throughput event đơn | **~2.450 req/s** |
| Max throughput batch (100 event/req) | **~107 req/s (~10.700 event/s)** |
| Max sustained rate (0 error, p95 < 2ms) | **200 req/s** |
| Buffer drain rate (lý thuyết) | **~16.7 event/s** |
