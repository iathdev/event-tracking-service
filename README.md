# Event Tracking Service

Service Go xử lý event tracking cho hệ thống Exam Platform (IELTS & TOEIC). Nhận event qua HTTP API, buffer vào Redis, batch insert vào PostgreSQL.

## Kiến trúc

```
Client (Browser)
    │  HTTP POST
    ▼
┌─────────────────────────────────────────────┐
│         event-tracking-service (Go)         │
│                                             │
│  Handler ──▶ Redis Buffer ──▶ Scheduler     │
│  (validate,     (RPUSH)      (batch insert) │
│   enrich)                         │         │
│                              PostgreSQL      │
└─────────────────────────────────────────────┘
```

- **Handler**: validate request, enrich metadata (IP, User-Agent), push vào Redis, trả 200 OK ngay
- **Scheduler**: chạy mỗi N giây, pop batch từ Redis, bulk insert vào PostgreSQL
- **Dead Letter**: event lỗi sau 3 lần retry được đẩy vào Redis dead letter list

## Yêu cầu

- Go 1.24+
- PostgreSQL 16
- Redis 7
- [golang-migrate](https://github.com/golang-migrate/migrate) (cho database migration)

## Cài đặt & Chạy

```bash
# 1. Khởi động PostgreSQL & Redis
docker-compose up -d

# 2. Cấu hình environment
cp .env.example .env

# 3. Cài đặt dependencies
make dep

# 4. Chạy migration
make migrate-up

# 5. Chạy service
make run
```

Service chạy tại `http://localhost:8080`.

## API Endpoints

| Method | Path | Auth | Mô tả |
|--------|------|------|-------|
| GET | `/health` | - | Health check |
| GET | `/docs` | - | API docs (Scalar UI) |
| POST | `/api/v1/events` | JWT | Gửi 1 event |
| POST | `/api/v1/events/batch` | JWT | Gửi batch event (max 100) |
| GET | `/api/v1/monitoring/queue-stats` | API Key | Thống kê queue |

### Ví dụ

```bash
# Gửi 1 event
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "event": "click_action_test_direction",
    "screen": "test_direction",
    "user_id": 456,
    "batch_id": 123,
    "properties": {
      "product_line": "IELTS",
      "action_test_direction": "continue"
    },
    "occurred_at": "2026-03-16T10:00:00Z"
  }'

# Response
{"success": true, "message": "Event accepted"}
```

## Makefile Commands

| Command | Mô tả |
|---------|-------|
| `make run` | Chạy local (macOS) |
| `make build` | Build binary cho Linux |
| `make dep` | `go mod tidy` |
| `make migrate-up` | Áp dụng migration |
| `make migrate-down` | Rollback 1 migration |
| `make migrate-create` | Tạo migration mới |

## Cấu hình

Cấu hình qua file `.env`. Xem `.env.example` để biết tất cả biến.

| Biến | Mặc định | Mô tả |
|------|----------|-------|
| `APP_PORT` | 8080 | Port HTTP server |
| `SCHEDULER_PROCESS_INTERVAL_SECONDS` | 60 | Chu kỳ batch process (giây) |
| `EVENT_BUFFER_BATCH_SIZE` | 1500 | Số event mỗi batch insert |
| `EVENT_BUFFER_MAX_RETRIES` | 3 | Số lần retry trước khi vào dead letter |
| `LOG_CHANNEL` | console | `console` hoặc `signoz` |
| `TRACING_ENABLE` | false | Bật OpenTelemetry tracing |
| `SENTRY_ENABLE` | false | Bật Sentry error tracking |

## Cấu trúc thư mục

```
├── main.go                    # Entrypoint, DI wiring
├── config/                    # Configuration loading
├── internal/
│   ├── handlers/              # HTTP handlers
│   ├── middleware/             # JWT auth, API key, CORS
│   ├── dtos/                  # Request/Response DTOs
│   ├── models/                # GORM models
│   ├── repositories/          # Database access
│   ├── services/              # EventBuffer, EventProcessor
│   ├── scheduler/             # Cron jobs (gocron + Redis lock)
│   └── httpserver/            # Server setup & routing
├── pkg/
│   ├── common/                # Response helpers, JSON types
│   ├── database/              # PostgreSQL & Redis connections
│   ├── observe/               # OTel, Sentry, logging middleware
│   └── logger/                # Structured logging (Zap)
├── migrations/                # SQL migrations (golang-migrate)
├── scripts/                   # k6 load test scripts
└── docs/                      # API spec, event list, test plans
```

## Tài liệu

- [API Documentation](http://localhost:8080/docs) — Scalar UI (khi service đang chạy)
- [docs/api/openapi.yaml](docs/api/openapi.yaml) — OpenAPI 3.0 spec
- [docs/plans/events.md](docs/plans/events.md) — Danh sách 28 events & properties
- [docs/api/curl-examples.md](docs/api/curl-examples.md) — Curl examples cho từng event
- [docs/plans/plan.md](docs/plans/plan.md) — Kế hoạch kiến trúc & implementation
- [docs/plans/plan-kafka.md](docs/plans/plan-kafka.md) — Kế hoạch Kafka edition
- [docs/testing/api-test-plan.md](docs/testing/api-test-plan.md) — Test cases & curl examples
- [docs/testing/load-test-plan.md](docs/testing/load-test-plan.md) — Kế hoạch load test
