# Implementation Plan: Event Tracking Service (Go)

> Xây dựng service Go lưu log event của hệ thống Exam Platform (IELTS & TOEIC)

---

## Tổng quan

### Hiện trạng
- **Go service skeleton** đã có sẵn: Gin, PostgreSQL (GORM), Redis, Scheduler (gocron), Observability (OTel, Sentry), JWT Auth
- **Laravel service** đang xử lý event tracking với Redis buffer → batch insert DB
- Cần xây service Go **độc lập** nhận event từ client, buffer qua Redis, batch insert PostgreSQL

### Mục tiêu
- Nhận event tracking từ FE (exam platform) qua HTTP API
- Buffer event vào Redis → batch processing vào PostgreSQL mỗi N phút
- Hỗ trợ cả TOEIC & IELTS, ưu tiên IELTS trước
- Thiết kế generic để mở rộng thêm event mới dễ dàng

---

## Danh sách Event cần tracking (Phase 1)

### Cấu trúc event chung

Mỗi event đều có các properties chung:

```json
{
  "event": "click_action_test_direction",
  "screen": "test_direction",
  "user_id": 456,
  "batch_id": 123,
  "batch_candidate_id": 789,
  "time_log": "2026-03-11T10:00:00Z",
  "properties": {
    "product_line_id": 1,
    "skill_id": 1,
    "action_test_direction": "continue"
  }
}
```

### Bảng event theo màn hình

| # | Screen | Event Name | Product Line | Skill | Event Properties |
|---|--------|-----------|-------------|-------|-----------------|
| 1 | Test Direction | `click_action_test_direction` | IELTS, TOEIC | 4 skills | `action_test_direction`: cancel/continue |
| 2 | Regulation | `click_agree_exam_regulation` | IELTS, TOEIC | 4 skills | `action_regulation`: agree |
| 3 | Check Audio | `click_continue_audio` | IELTS, TOEIC | Listening | - |
| 4 | Test Room | `join_test` | IELTS, TOEIC | 4 skills | - |
| 5 | Test Room | `exit_test` | IELTS, TOEIC | 4 skills | - |
| 6 | Test Room | `change_part` | IELTS, TOEIC | 4 skills | `part_id`, `from_part`, `to_part` |
| 7 | Test Room | `change_question` | IELTS, TOEIC | 4 skills | `question_id`, `position` |
| 8 | Test Room | `start_skill` | IELTS, TOEIC | 4 skills | `skill_id` |
| 9 | Test Room | `submit_skill` | IELTS, TOEIC | 4 skills | `skill_id` |
| 10 | Test Room | `focus_page` | IELTS, TOEIC | 4 skills | - |
| 11 | Test Room | `un_focus_page` | IELTS, TOEIC | 4 skills | - |
| 12 | Test Room | `over_timer` | IELTS, TOEIC | 4 skills | - |
| 13 | Test Room | `network_offline` | IELTS, TOEIC | 4 skills | - |
| 14 | Test Room | `submit_by_admin` | IELTS, TOEIC | 4 skills | - |

---

## Kiến trúc tổng quan

```
┌──────────┐     ┌─────────────────────────────────────────────────┐
│  Client   │     │        event-tracking-service (Go)              │
│ (Browser) │────▶│                                                 │
└──────────┘     │  ┌───────────┐   ┌──────────┐   ┌───────────┐  │
     HTTP POST   │  │  Handler   │──▶│  Redis    │──▶│ Scheduler │  │
                 │  │ (validate) │   │  Buffer   │   │ (batch    │  │
                 │  └───────────┘   └──────────┘   │  insert)  │  │
                 │                                   └─────┬─────┘  │
                 │                                         │        │
                 │                                   ┌─────▼─────┐  │
                 │                                   │ PostgreSQL │  │
                 │                                   └───────────┘  │
                 └─────────────────────────────────────────────────┘
```

---

## Plan theo Phase

---

### Phase 1: Database & Models (Ưu tiên cao)

#### 1.1 Tạo migration PostgreSQL

**File**: `migrations/001_create_tracking_events_table.sql`

```sql
CREATE TABLE IF NOT EXISTS tracking_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event         VARCHAR(100) NOT NULL,          -- Tên event (vd: click_action_test_direction, join_test)
    screen        VARCHAR(100) NOT NULL,          -- Màn hình phát sinh event: test_direction, regulation, test_room, ...
    user_id       BIGINT       NOT NULL,          -- ID user thực hiện event
    batch_id      BIGINT,                         -- ID batch (nullable, một số event không có batch)
    batch_candidate_id BIGINT,                    -- ID batch_candidate (nullable, gắn thí sinh với đợt thi)
    properties    JSONB        DEFAULT '{}',      -- Dữ liệu dynamic theo từng loại event (product_line_id, skill_id, submission_id, ...)
    meta_data     JSONB        DEFAULT '{}',      -- Metadata bổ sung từ request: IP, UserAgent, Device, OS, Browser...
    time_log      TIMESTAMPTZ  NOT NULL,          -- Thời điểm event xảy ra phía client (ISO 8601)
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()  -- Thời điểm record được tạo trong DB
);

CREATE INDEX idx_tracking_events_event ON tracking_events (event);
CREATE INDEX idx_tracking_events_screen ON tracking_events (screen);
CREATE INDEX idx_tracking_events_user_id ON tracking_events (user_id);
CREATE INDEX idx_tracking_events_batch_id ON tracking_events (batch_id);
CREATE INDEX idx_tracking_events_batch_candidate_id ON tracking_events (batch_candidate_id);
CREATE INDEX idx_tracking_events_time_log ON tracking_events (time_log);
```

#### 1.2 GORM Model

**File**: `internal/models/tracking_event.go`

```go
type TrackingEvent struct {
    models.BaseModel
    Event      string         `gorm:"column:event;type:varchar(100);not null" json:"event"`
    Screen     string         `gorm:"column:screen;type:varchar(100);not null" json:"screen"`
    UserID     int64          `gorm:"column:user_id;not null" json:"user_id"`
    BatchID           *int64         `gorm:"column:batch_id" json:"batch_id,omitempty"`                    // nullable
    BatchCandidateID  *int64         `gorm:"column:batch_candidate_id" json:"batch_candidate_id,omitempty"` // nullable
    Properties        common.JSONMap `gorm:"column:properties;type:jsonb;default:'{}'" json:"properties"`
    MetaData   common.JSONMap `gorm:"column:meta_data;type:jsonb;default:'{}'" json:"meta_data"`
    TimeLog    time.Time      `gorm:"column:time_log;type:timestamptz;not null" json:"time_log"`
    CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (TrackingEvent) TableName() string {
    return "tracking_events"
}
```

> **Lưu ý**: `user_id` và `batch_id` là dedicated columns (query thường xuyên, native B-tree index). Các dynamic fields (`product_line_id`, `skill_id`, `submission_id`,...) nằm trong `properties` JSONB — không cần migrate DB khi thêm field mới.

---

### Phase 2: Redis Event Buffer

#### 2.1 Event Buffer Service

**File**: `internal/service/event_buffer.go`

Chức năng:
- `Push(event)` — serialize event → RPUSH vào Redis list `event_tracking:events`
- `PopBatch(batchSize)` — LPOP N events từ Redis list
- `QueueSize()` — LLEN kiểm tra queue length
- `PushDeadLetter(event)` — đẩy event lỗi vào `event_tracking:dead_letter`

```go
type EventBuffer struct {
    redis     *redis.Client
    logger    *zap.Logger
    queueKey  string          // "event_tracking:events"
    deadKey   string          // "event_tracking:dead_letter"
    batchSize int
}

func (b *EventBuffer) Push(ctx context.Context, event *TrackingEventDTO) error
func (b *EventBuffer) PopBatch(ctx context.Context, size int) ([]TrackingEventDTO, error)
func (b *EventBuffer) QueueSize(ctx context.Context) (int64, error)
func (b *EventBuffer) PushDeadLetter(ctx context.Context, event *TrackingEventDTO, reason string) error
```

#### 2.2 Config bổ sung

Thêm vào `config/config.go`:

```go
type EventBufferConfig struct {
    QueueKey      string        // "event_tracking:events"
    DeadLetterKey string        // "event_tracking:dead_letter"
    BatchSize     int           // 1500
    MaxRetries    int           // 3
}
```

Env vars:
```
EVENT_BUFFER_QUEUE_KEY=event_tracking:events
EVENT_BUFFER_DEAD_LETTER_KEY=event_tracking:dead_letter
EVENT_BUFFER_BATCH_SIZE=1500
EVENT_BUFFER_MAX_RETRIES=3
```

---

### Phase 3: API Handler (Nhận event từ client)

#### 3.1 Request DTO

**File**: `internal/dto/tracking_event_dto.go`

```go
type CreateTrackingEventRequest struct {
    Event      string                 `json:"event" binding:"required"`
    Screen     string                 `json:"screen" binding:"required"`
    UserID     int64                  `json:"user_id" binding:"required"`
    BatchID          *int64                 `json:"batch_id,omitempty"`
    BatchCandidateID *int64                 `json:"batch_candidate_id,omitempty"`
    Properties       map[string]interface{} `json:"properties,omitempty"`    // dynamic fields: product_line_id, skill_id, submission_id, ...
    TimeLog    *string                `json:"time_log,omitempty"`      // ISO 8601, default = now()
}

// Batch request - client gửi nhiều events 1 lần
type CreateBatchTrackingEventRequest struct {
    Events []CreateTrackingEventRequest `json:"events" binding:"required,min=1,max=100,dive"`
}
```

#### 3.2 Handler

**File**: `internal/handlers/tracking_event_handler.go`

```go
type TrackingEventHandler struct {
    buffer *service.EventBuffer
    logger *zap.Logger
}

// POST /api/v1/events         — single event
// POST /api/v1/events/batch   — batch events (tối đa 100)
```

Luồng xử lý:
1. Validate request (binding)
2. Enrich metadata: gom IP, User-Agent, Device từ request header vào `meta_data` JSONB
3. Push vào Redis buffer (không ghi DB trực tiếp)
4. Trả 202 Accepted (async processing)

#### 3.3 Routes

**File**: cập nhật `internal/httpserver/http_start.go`

```go
api := router.Group("/api/v1/")
{
    events := api.Group("/events")
    {
        events.POST("", trackingEventHandler.Create)
        events.POST("/batch", trackingEventHandler.CreateBatch)
    }
}
```

---

### Phase 4: Batch Processor (Scheduler Job)

#### 4.1 Processor Service

**File**: `internal/service/event_processor.go`

```go
type EventProcessor struct {
    buffer *EventBuffer
    repo   *repository.TrackingEventRepository
    logger *zap.Logger
    cfg    *config.EventBufferConfig
}

func (p *EventProcessor) ProcessQueue(ctx context.Context) error {
    // 1. Kiểm tra queue size
    // 2. Pop batch từ Redis
    // 3. Transform DTO → Model
    // 4. Bulk INSERT vào PostgreSQL (GORM CreateInBatches)
    // 5. Nếu lỗi → retry (max 3 lần) → dead letter
    // 6. Log metrics (processed count, duration)
}
```

#### 4.2 Register Scheduler Job

**File**: cập nhật `internal/scheduler/scheduler.go`

```go
func (s *Scheduler) RegisterJobs() error {
    _, err := s.scheduler.NewJob(
        gocron.DurationJob(s.cfg.Scheduler.ProcessInterval),  // mỗi 2 phút
        gocron.NewTask(s.processEventQueue),
        gocron.WithName("process_event_queue"),
    )
    return err
}

func (s *Scheduler) processEventQueue() {
    // Gọi EventProcessor.ProcessQueue()
}
```

---

### Phase 5: Repository Layer

#### 5.1 Tracking Event Repository

**File**: `internal/repository/tracking_event_repository.go`

```go
type TrackingEventRepository struct {
    db *gorm.DB
}

func (r *TrackingEventRepository) BulkInsert(ctx context.Context, events []models.TrackingEvent) error {
    return r.db.WithContext(ctx).CreateInBatches(events, 500).Error
}

func (r *TrackingEventRepository) GetByUserID(ctx context.Context, userID int64) ([]models.TrackingEvent, error)
func (r *TrackingEventRepository) GetByBatchID(ctx context.Context, batchID int64) ([]models.TrackingEvent, error)
func (r *TrackingEventRepository) GetByBatchCandidateID(ctx context.Context, batchCandidateID int64) ([]models.TrackingEvent, error)
func (r *TrackingEventRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error)
```

---

### Phase 6: Monitoring & Health

#### 6.1 Queue Stats Endpoint

**File**: `internal/handlers/monitoring_handler.go`

```go
// GET /api/v1/monitoring/queue-stats (internal API key auth)
// Response:
{
    "queue_size": 1234,
    "dead_letter_size": 5,
    "last_processed_at": "2026-03-11T10:00:00Z"
}
```

#### 6.2 Health check mở rộng

Cập nhật health check để kiểm tra Redis + PostgreSQL connectivity.

---

## Cấu trúc thư mục sau khi implement

```
event-tracking-service/
├── main.go
├── config/
│   └── config.go                          # + EventBufferConfig
├── migrations/
│   └── 001_create_tracking_events_table.sql
├── internal/
│   ├── models/
│   │   ├── base_model.go                  # (existing)
│   │   └── tracking_event.go              # NEW
│   ├── dto/
│   │   └── tracking_event_dto.go          # NEW
│   ├── repository/
│   │   └── tracking_event_repository.go   # NEW
│   ├── service/
│   │   ├── event_buffer.go                # NEW - Redis buffer
│   │   └── event_processor.go             # NEW - Batch processor
│   ├── handlers/
│   │   ├── healthcheck_handler.go         # (existing)
│   │   ├── tracking_event_handler.go      # NEW
│   │   └── monitoring_handler.go          # NEW
│   ├── middleware/
│   │   ├── auth.go                        # (existing)
│   │   └── api-key-internal.go            # (existing)
│   ├── httpserver/
│   │   ├── httpserver.go                  # (existing)
│   │   └── http_start.go                 # UPDATE - add routes
│   └── scheduler/
│       └── scheduler.go                   # UPDATE - register job
├── pkg/                                    # (existing - no changes)
│   ├── common/
│   ├── database/
│   ├── logger/
│   └── observe/
└── docs/
    └── openapi.yaml                        # NEW - API spec
```

---

## Thứ tự implement (Step by step)

```
Step 1: Migration + Model
   └── tracking_event.go + SQL migration
   └── Auto-migrate hoặc golang-migrate

Step 2: DTO + Validation
   └── tracking_event_dto.go
   └── Request/Response structs

Step 3: Redis Event Buffer
   └── event_buffer.go
   └── Push/Pop/QueueSize methods
   └── Config mới cho buffer

Step 4: Repository
   └── tracking_event_repository.go
   └── BulkInsert, query methods

Step 5: API Handler + Routes
   └── tracking_event_handler.go
   └── POST /api/v1/events
   └── POST /api/v1/events/batch
   └── Cập nhật http_start.go

Step 6: Batch Processor + Scheduler
   └── event_processor.go
   └── Cập nhật scheduler.go
   └── Register cron job

Step 7: Monitoring
   └── monitoring_handler.go
   └── Queue stats endpoint

Step 8: Testing
   └── Unit tests cho buffer, processor, handler
   └── Integration test với Redis + PostgreSQL
```

---

## API Spec tóm tắt

### POST /api/v1/events

Gửi 1 event.

**Request:**
```json
{
  "event": "click_action_test_direction",
  "screen": "test_direction",
  "user_id": 456,
  "batch_id": 123,
  "batch_candidate_id": 789,
  "properties": {
    "product_line_id": 1,
    "action_test_direction": "continue"
  },
  "time_log": "2026-03-11T10:00:00Z"
}
```

**Response:** `202 Accepted`
```json
{
  "message": "Event accepted",
  "data": null
}
```

### POST /api/v1/events/batch

Gửi nhiều events (max 100).

**Request:**
```json
{
  "events": [
    {
      "event": "click_action_test_direction",
      "screen": "test_direction",
      "user_id": 456,
      "batch_id": 123,
      "batch_candidate_id": 789,
      "properties": { "product_line_id": 1, "action_test_direction": "continue" },
      "time_log": "2026-03-11T10:00:00Z"
    },
    {
      "event": "click_agree_exam_regulation",
      "screen": "regulation",
      "user_id": 456,
      "batch_id": 123,
      "batch_candidate_id": 789,
      "properties": { "product_line_id": 1, "action_regulation": "agree" },
      "time_log": "2026-03-11T10:00:05Z"
    }
  ]
}
```

**Response:** `202 Accepted`
```json
{
  "message": "2 events accepted",
  "data": null
}
```

### GET /api/v1/monitoring/queue-stats (Internal)

**Response:**
```json
{
  "message": "ok",
  "data": {
    "queue_size": 1234,
    "dead_letter_size": 5
  }
}
```

---

## So sánh Laravel hiện tại vs Go mới

| Aspect | Laravel (hiện tại) | Go (mới) |
|--------|-------------------|----------|
| Database | MySQL | PostgreSQL |
| Buffer | Redis (RPUSH/LPOP) | Redis (RPUSH/LPOP) — giống |
| Batch size | 1,500 | 1,500 (configurable) |
| Schedule | Mỗi 3 phút | Mỗi 2 phút (configurable) |
| Auth | Sanctum (session token) | JWT Bearer token |
| Event format | CandidateLogData DTO | Unified TrackingEventDTO |
| Locking | Redis distributed lock | gocron-redis-lock — giống |
| Dead letter | Có | Có |
| Fallback DB write | Có | Có (Phase 2) |

---

## Lưu ý quan trọng

1. **Hybrid design**: `user_id`, `batch_id`, `batch_candidate_id` là dedicated columns (query thường xuyên, native index). Các dynamic fields vẫn nằm trong `properties JSONB` — không cần thay đổi schema khi thêm event mới
2. **Backward compatible**: Service Go chạy song song với Laravel, không cần migrate data cũ
3. **Performance**: Go + Redis buffer đảm bảo latency thấp cho client (202 Accepted ngay)
4. **Scalability**: Có thể scale horizontal scheduler bằng distributed lock
5. **Observability**: Đã có sẵn OTel + Sentry trong skeleton
