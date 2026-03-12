# Kế hoạch kiểm thử chức năng API

## 1. Phạm vi

Kiểm thử chức năng HTTP API của Event Tracking Service, bao gồm:

- Tạo event (đơn lẻ & theo batch)
- Xác thực & phân quyền (JWT, API key)
- Validate request & xử lý lỗi
- Endpoint giám sát thống kê queue
- Endpoint health check

**Ngoài phạm vi:** Nội bộ database, xử lý scheduler, hạ tầng.

## 2. Chiến lược

| Cấp độ | Đối tượng | Phương pháp |
|--------|-----------|-------------|
| API Contract | Định dạng request/response, status code | Test HTTP tự động |
| Xác thực | JWT & API key middleware | Kịch bản xác thực positive/negative |
| Validation | Trường bắt buộc, kiểu dữ liệu, giới hạn | Test biên & input không hợp lệ |
| Xử lý lỗi | Payload sai định dạng, lỗi server | Fault injection |
| Tích hợp | Luồng HTTP → Redis buffer | End-to-end với Redis thật |

## 3. Môi trường

| Thành phần | Chi tiết |
|------------|----------|
| App | `localhost:8080` (mặc định) |
| PostgreSQL | Local hoặc Docker, đã chạy migration |
| Redis | Local hoặc Docker, port mặc định 6379 |
| Xác thực | JWT hợp lệ ký bằng test secret |
| API Key | `API_KEY_INTERNAL` đặt trong env |

## 4. Dữ liệu test

**Event đơn hợp lệ:**
```json
{
  "event": "page_view",
  "screen": "dashboard",
  "user_id": 12345,
  "batch_id": 1,
  "properties": {"section": "overview"},
  "meta_data": {"source": "web"},
  "occurred_at": "2026-03-15T10:00:00Z"
}
```

**Batch hợp lệ (2 event):**
```json
{
  "events": [
    {"event": "click", "screen": "exam_list", "user_id": 1},
    {"event": "submit", "screen": "exam_detail", "user_id": 2}
  ]
}
```

## 5. Các test case

### 5.1 Event đơn — `POST /api/v1/events`

| ID | Test case | Input | Kết quả mong đợi | Trạng thái |
|----|-----------|-------|-------------------|------------|
| E01 | Event hợp lệ (đầy đủ trường) | Payload đầy đủ tất cả trường | 200, `"Event accepted"` | |
| E02 | Event hợp lệ (chỉ trường bắt buộc) | Chỉ `event`, `screen`, `user_id` | 200, `"Event accepted"` | |
| E03 | Thiếu trường `event` | Bỏ `event` | 422, lỗi validation | |
| E04 | Thiếu trường `screen` | Bỏ `screen` | 422, lỗi validation | |
| E05 | Thiếu trường `user_id` | Bỏ `user_id` | 422, lỗi validation | |
| E06 | `user_id` = 0 | `"user_id": 0` | 422, required fail với giá trị zero | |
| E07 | `occurred_at` sai định dạng | `"occurred_at": "not-a-date"` | 200 (parse khi xử lý) hoặc lỗi | |
| E08 | `properties` rỗng | `"properties": {}` | 200 | |
| E09 | `properties` lồng nhau | JSON map lồng sâu | 200 | |
| E10 | Tự động bổ sung metadata | Bất kỳ event hợp lệ | `meta_data` chứa `ip`, `user_agent` | |
| E11 | Body request rỗng | `{}` | 422 | |
| E12 | JSON sai định dạng | `{broken` | 422 | |
| E13 | `user_id` âm | `"user_id": -1` | 200 (không có constraint) hoặc 422 | |
| E14 | Chuỗi `event` rất dài | 200+ ký tự | 200 (bị cắt ở DB) hoặc lỗi | |

### 5.2 Batch Event — `POST /api/v1/events/batch`

| ID | Test case | Input | Kết quả mong đợi | Trạng thái |
|----|-----------|-------|-------------------|------------|
| B01 | Batch hợp lệ (2 event) | 2 event hợp lệ | 200, `"2 events accepted"` | |
| B02 | Batch 1 event | 1 event hợp lệ | 200, `"1 events accepted"` | |
| B03 | Batch tối đa (100 event) | 100 event hợp lệ | 200, `"100 events accepted"` | |
| B04 | Vượt quá tối đa (101 event) | 101 event | 422, lỗi validation max | |
| B05 | Mảng events rỗng | `{"events": []}` | 422, lỗi validation min=1 | |
| B06 | Thiếu trường `events` | `{}` | 422 | |
| B07 | Một event không hợp lệ trong batch | 99 hợp lệ + 1 thiếu `event` | 422 (dive validation) | |
| B08 | Event trùng lặp | Cùng payload lặp lại | 200, tất cả được chấp nhận | |

### 5.3 Xác thực

| ID | Test case | Input | Kết quả mong đợi | Trạng thái |
|----|-----------|-------|-------------------|------------|
| A01 | JWT hợp lệ | `Authorization: Bearer <valid>` | 200 | |
| A02 | Thiếu header Authorization | Không có header | 401, `"Unauthorized"` | |
| A03 | Token không hợp lệ | `Bearer invalid123` | 401 | |
| A04 | Token hết hạn | JWT đã expired | 401 | |
| A05 | Sai signing method | Token ký bằng RSA | 401 | |
| A06 | Thiếu prefix Bearer | `Authorization: <token>` | 401, `"token invalid"` | |
| A07 | Bearer token rỗng | `Authorization: Bearer ` | 401 | |
| A08 | API key nội bộ hợp lệ | `x-api-key: <correct>` trên monitoring | 200 | |
| A09 | API key không hợp lệ | `x-api-key: wrong` trên monitoring | 401, `"API KEY invalid"` | |
| A10 | Thiếu API key | Không có header trên monitoring | 401 | |

### 5.4 Giám sát — `GET /api/v1/monitoring/queue-stats`

| ID | Test case | Input | Kết quả mong đợi | Trạng thái |
|----|-----------|-------|-------------------|------------|
| M01 | Thống kê queue với key hợp lệ | `x-api-key` hợp lệ | 200, kích thước queue & dead letter | |
| M02 | Thống kê queue không có key | Không xác thực | 401 | |

### 5.5 Health Check — `GET /health`

| ID | Test case | Input | Kết quả mong đợi | Trạng thái |
|----|-----------|-------|-------------------|------------|
| H01 | Health check | GET request | 200 | |

## 6. Trường hợp biên (Edge Cases)

| ID | Kịch bản | Hành vi mong đợi |
|----|----------|-------------------|
| EC01 | Gửi event đồng thời (100+) | Tất cả trả về 200, không mất dữ liệu |
| EC02 | Redis ngừng hoạt động khi push event | 500 Internal Server Error |
| EC03 | Payload `properties` rất lớn (1MB+) | Được chấp nhận hoặc lỗi body limit của Gin |
| EC04 | Unicode/emoji trong tên event | 200, lưu trữ đúng |
| EC05 | SQL injection trong các trường string | Lưu trữ an toàn (GORM parameterized) |
| EC06 | XSS payload trong properties | Lưu trữ nguyên trạng (không render) |
| EC07 | Giá trị null ở các trường tùy chọn | 200, lưu trữ dạng null |
| EC08 | Content-Type không phải application/json | 422 hoặc 400 |
| EC09 | Gửi trùng lặp liên tục nhanh | Tất cả được chấp nhận (không dedup) |
| EC10 | Request có thêm trường không xác định | 200, trường thừa bị bỏ qua |

## 7. Kiểm tra response lỗi

Tất cả response lỗi phải tuân theo cấu trúc:

```json
{
  "message": "<mô tả lỗi>",
  "error": "<chi tiết lỗi hoặc null>",
  "data": null
}
```

| Status Code | Khi nào |
|-------------|---------|
| 200 | Event được chấp nhận |
| 400 | Bad request chung |
| 401 | Lỗi xác thực (JWT hoặc API key) |
| 422 | Lỗi validation (thiếu/sai trường) |
| 500 | Lỗi nội bộ (Redis ngừng, v.v.) |

## 8. Ví dụ curl

> Thay `$TOKEN` bằng JWT token hợp lệ, `$API_KEY` bằng giá trị `API_KEY_INTERNAL`.

---

### 8.1 Health Check (H01)

```bash
curl -s http://localhost:8080/health
```

**Response — 200 OK:**
```json
{"message":"ok"}
```

---

### 8.2 Event đơn — đầy đủ trường (E01)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"event":"page_view","screen":"dashboard","user_id":12345,"batch_id":1,"properties":{"section":"overview"},"meta_data":{"source":"web"},"occurred_at":"2026-03-15T10:00:00Z"}'
```

**Response — 200 OK:**
```json
{"message":"Event accepted"}
```

---

### 8.3 Event đơn — chỉ trường bắt buộc (E02)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"event":"click","screen":"exam_list","user_id":1}'
```

**Response — 200 OK:**
```json
{"message":"Event accepted"}
```

---

### 8.4 Thiếu trường `event` (E03)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"screen":"dashboard","user_id":12345}'
```

**Response — 422 Unprocessable Entity:**
```json
{"message":"Validation failed","error":"Key: 'CreateTrackingEventRequest.Event' Error:Field validation for 'Event' failed on the 'required' tag"}
```

---

### 8.5 Thiếu trường `screen` (E04)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"event":"page_view","user_id":12345}'
```

**Response — 422 Unprocessable Entity:**
```json
{"message":"Validation failed","error":"Key: 'CreateTrackingEventRequest.Screen' Error:Field validation for 'Screen' failed on the 'required' tag"}
```

---

### 8.6 Thiếu trường `user_id` (E05)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"event":"page_view","screen":"dashboard"}'
```

**Response — 422 Unprocessable Entity:**
```json
{"message":"Validation failed","error":"Key: 'CreateTrackingEventRequest.UserID' Error:Field validation for 'UserID' failed on the 'required' tag"}
```

---

### 8.7 `user_id` = 0 (E06)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"event":"page_view","screen":"dashboard","user_id":0}'
```

**Response — 422 Unprocessable Entity:**
```json
{"message":"Validation failed","error":"Key: 'CreateTrackingEventRequest.UserID' Error:Field validation for 'UserID' failed on the 'required' tag"}
```

---

### 8.8 JSON sai định dạng (E12)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{broken'
```

**Response — 422 Unprocessable Entity:**
```json
{"message":"Validation failed","error":"invalid character 'b' looking for beginning of object key string"}
```

---

### 8.9 Body request rỗng (E11)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{}'
```

**Response — 422 Unprocessable Entity:**
```json
{"message":"Validation failed","error":"Key: 'CreateTrackingEventRequest.Event' Error:Field validation for 'Event' failed on the 'required' tag\nKey: 'CreateTrackingEventRequest.Screen' Error:Field validation for 'Screen' failed on the 'required' tag\nKey: 'CreateTrackingEventRequest.UserID' Error:Field validation for 'UserID' failed on the 'required' tag"}
```

---

### 8.10 Batch hợp lệ (B01)

```bash
curl -s -X POST http://localhost:8080/api/v1/events/batch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"events":[{"event":"click","screen":"exam_list","user_id":1},{"event":"submit","screen":"exam_detail","user_id":2}]}'
```

**Response — 200 OK:**
```json
{"message":"2 events accepted"}
```

---

### 8.11 Batch rỗng (B05)

```bash
curl -s -X POST http://localhost:8080/api/v1/events/batch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"events":[]}'
```

**Response — 422 Unprocessable Entity:**
```json
{"message":"Validation failed","error":"Key: 'CreateBatchTrackingEventRequest.Events' Error:Field validation for 'Events' failed on the 'min' tag"}
```

---

### 8.12 Batch — một event thiếu trường (B07)

```bash
curl -s -X POST http://localhost:8080/api/v1/events/batch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"events":[{"event":"click","screen":"exam_list","user_id":1},{"screen":"exam_detail","user_id":2}]}'
```

**Response — 422 Unprocessable Entity:**
```json
{"message":"Validation failed","error":"Key: 'CreateBatchTrackingEventRequest.Events[1].Event' Error:Field validation for 'Event' failed on the 'required' tag"}
```

---

### 8.13 Thiếu header Authorization (A02)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{"event":"click","screen":"test","user_id":1}'
```

**Response — 401 Unauthorized:**
```json
{"message":"Unauthorized","error":"token invalid"}
```

---

### 8.14 Token không hợp lệ (A03)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer invalid_token_123" \
  -d '{"event":"click","screen":"test","user_id":1}'
```

**Response — 401 Unauthorized:**
```json
{"message":"Unauthorized","error":"token is malformed: token contains an invalid number of segments"}
```

---

### 8.15 Thiếu prefix Bearer (A06)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: some_token_without_bearer" \
  -d '{"event":"click","screen":"test","user_id":1}'
```

**Response — 401 Unauthorized:**
```json
{"message":"Unauthorized","error":"token invalid"}
```

---

### 8.16 Thống kê queue — hợp lệ (M01)

```bash
curl -s http://localhost:8080/api/v1/monitoring/queue-stats \
  -H "x-api-key: $API_KEY"
```

**Response — 200 OK:**
```json
{"message":"ok","data":{"queue_size":5,"dead_letter_size":0}}
```

---

### 8.17 Thống kê queue — thiếu API key (M02)

```bash
curl -s http://localhost:8080/api/v1/monitoring/queue-stats
```

**Response — 401 Unauthorized:**
```json
{"message":"API KEY invalid"}
```

---

### 8.18 API key không hợp lệ (A09)

```bash
curl -s http://localhost:8080/api/v1/monitoring/queue-stats \
  -H "x-api-key: wrong_key"
```

**Response — 401 Unauthorized:**
```json
{"message":"API KEY invalid"}
```

---

### 8.19 Unicode/emoji trong tên event (EC04)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"event":"thi_IELTS_\ud83c\udf93","screen":"man_hinh_chinh","user_id":1}'
```

**Response — 200 OK:**
```json
{"message":"Event accepted"}
```

---

### 8.20 SQL injection (EC05)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"event":"test OR 1=1; DROP TABLE tracking_events;--","screen":"test","user_id":1}'
```

**Response — 200 OK (an toàn, GORM parameterized query):**
```json
{"message":"Event accepted"}
```

---

### 8.21 Sai Content-Type (EC08)

```bash
curl -s -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: text/plain" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"event":"click","screen":"test","user_id":1}'
```

**Response — 422 Unprocessable Entity:**
```json
{"message":"Validation failed","error":"Key: 'CreateTrackingEventRequest.Event' Error:Field validation for 'Event' failed on the 'required' tag\nKey: 'CreateTrackingEventRequest.Screen' Error:Field validation for 'Screen' failed on the 'required' tag\nKey: 'CreateTrackingEventRequest.UserID' Error:Field validation for 'UserID' failed on the 'required' tag"}
```

## 9. Các bước thực thi

1. **Cài đặt:** Khởi động PostgreSQL, Redis, chạy migration, khởi động service
2. **Cài đặt xác thực:** Tạo JWT token hợp lệ/không hợp lệ/hết hạn
3. **Chạy test:** Thực thi bộ test với service đang chạy
4. **Kiểm tra Redis:** Xác nhận event được đưa vào buffer `event_tracking:events`
5. **Xem kết quả:** Kiểm tra pass/fail, ghi nhận defect

## 10. Công cụ

| Công cụ | Mục đích |
|---------|----------|
| **Postman / Bruno** | Test API thủ công, chia sẻ collection |
| **Go `httptest`** | Unit/integration test bằng Go |
| **curl** | Gọi API nhanh |
| **Docker Compose** | Khởi tạo PostgreSQL + Redis |
| **redis-cli** | Kiểm tra buffer queue trực tiếp |
| **GitHub Actions / GitLab CI** | Thực thi test tự động |
