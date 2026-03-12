# Implementation Plan: Event Tracking Service — Kafka Edition

> Thay thế Redis Buffer bằng Kafka cho service Go lưu log event của hệ thống Exam Platform (IELTS & TOEIC)

---

## Context

Service hiện tại được thiết kế với Redis làm buffer (RPUSH/LPOP) + scheduler poll mỗi 2 phút để batch insert vào PostgreSQL. Chuyển sang Kafka để:
- Xử lý gần real-time (5s thay vì 2 phút polling)
- Durability tốt hơn (Kafka replication vs Redis AOF)
- Horizontal scaling tự nhiên (consumer group vs distributed lock)
- Replay capability khi cần reprocess
- Xử lý backpressure tốt hơn trong giờ thi cao điểm

---

## Kiến trúc mới

```
┌──────────┐     ┌──────────────────────────────────────────────────────────┐
│  Client   │     │          event-tracking-service (Go)                     │
│ (Browser) │────▶│                                                          │
└──────────┘     │  ┌───────────┐   ┌──────────┐   ┌───────────────────┐   │
     HTTP POST   │  │  Handler   │──▶│  Kafka    │──▶│ Consumer Group    │   │
                 │  │ (validate  │   │  Producer │   │ (batch accumulate │   │
                 │  │  + enrich) │   └──────────┘   │  + flush)         │   │
                 │  └───────────┘                    └────────┬──────────┘   │
                 │                                            │              │
                 │                                      ┌─────▼─────┐       │
                 │                                      │ PostgreSQL │       │
                 │                                      └───────────┘       │
                 │                                                          │
                 │  Lỗi sau 3 retry ──▶ DLQ topic                          │
                 └──────────────────────────────────────────────────────────┘
```

**Request flow:**
```
HTTP POST /api/v1/events → Handler (validate + enrich metadata)
  → Kafka Producer (topic: event-tracking.events)
  → Kafka Consumer Group (batch accumulate)
  → PostgreSQL bulk insert (GORM CreateInBatches)
  → Nếu lỗi sau 3 retry → produce to DLQ topic
```

**Thay đổi chính so với plan Redis:**
- Kafka Producer thay thế Redis RPUSH
- Kafka Consumer thay thế Scheduler + Redis LPOP
- Không cần gocron scheduler và Redis distributed lock cho event processing nữa

---

## Kafka Library: `segmentio/kafka-go`

Chọn vì:
- Pure Go, không cần CGo → giữ Dockerfile đơn giản (scratch image)
- API idiomatic Go với `context.Context`
- Built-in consumer group rebalancing + batch reading
- Có OTel instrumentation

---

## Topic Strategy

| Topic | Partitions | Replication Factor | Retention |
|-------|-----------|-------------------|-----------|
| `event-tracking.events` | 6 | 3 | 24h |
| `event-tracking.events.dlq` | 3 | 3 | 7d |

- **Partition key**: `user_id` → đảm bảo ordering events của cùng 1 user
- **Retention 24h**: safety net để replay nếu cần reprocess

---

## Consumer Batch Pattern

```
FetchMessage loop → accumulate buffer ([]TrackingEvent)
  → flush khi: buffer đạt 1500 HOẶC 5 giây trôi qua
  → GORM CreateInBatches(500) → CommitMessages
  → Lỗi: retry 3 lần (exponential backoff 1s, 2s, 4s) → DLQ
```

Manual offset commit — chỉ commit sau khi DB write thành công (at-least-once delivery).

---

## Danh sách Event cần tracking

### Cấu trúc event chung

```json
{
  "event": "click_action_test_direction",
  "screen": "test_direction",
  "user_id": 456,
  "batch_id": 123,
  "occurred_at": "2026-03-11T10:00:00Z",
  "properties": {
    "product_line": "IELTS",
    "skill": "Listening",
    "action_test_direction": "continue"
  }
}
```

### Events từ plan cũ (giữ nguyên)

| # | Screen | Event Name | Product Line | Skill | Event Properties |
|---|--------|-----------|-------------|-------|-----------------|
| 1 | Test Direction | `click_action_test_direction` | IELTS, TOEIC | 4 skills | `action_test_direction`: cancel/continue |
| 2 | Regulation | `click_agree_exam_regulation` | IELTS, TOEIC | 4 skills | `action_regulation`: agree |
| 3 | Check Audio | `click_continue_audio` | IELTS, TOEIC | Listening | - |
| 4 | Test Room | `join_test` | IELTS, TOEIC | 4 skills | - |
| 5 | Test Room | `exit_test` | IELTS, TOEIC | 4 skills | - |
| 6 | Test Room | `change_part` | IELTS, TOEIC | 4 skills | `part_id`, `from_part`, `to_part` |
| 7 | Test Room | `change_question` | IELTS, TOEIC | 4 skills | `question_id`, `position` |
| 8 | Test Room | `start_skill` | IELTS, TOEIC | 4 skills | `skill` |
| 9 | Test Room | `submit_skill` | IELTS, TOEIC | 4 skills | `skill` |
| 10 | Test Room | `focus_page` | IELTS, TOEIC | 4 skills | - |
| 11 | Test Room | `un_focus_page` | IELTS, TOEIC | 4 skills | - |
| 12 | Test Room | `over_timer` | IELTS, TOEIC | 4 skills | - |
| 13 | Test Room | `network_offline` | IELTS, TOEIC | 4 skills | - |
| 14 | Test Room | `submit_by_admin` | IELTS, TOEIC | 4 skills | - |

### Events bổ sung (mới)

| # | Event | Screen | Product Line | Properties đặc trưng |
|---|-------|--------|-------------|---------------------|
| 15 | `log_in_success` | Login | All | `user_id` |
| 16 | `click_action_your_test` | Your tests | All | `action_your_test`, `batch_id` |
| 17 | `note_question_ielts` | Test Room | IELTS | `question_id`, `submission_skill_id` |
| 18 | `highlight_question_ielts` | Test Room | IELTS | `question_id`, `submission_skill_id` |
| 19 | `view_note_ielts` | Test Room | IELTS | `question_id`, `submission_skill_id` |
| 20 | `delete_note_ielts` | Test Room | IELTS | `question_id`, `submission_skill_id` |
| 21 | `do_writing_test_toiec` | Test Room | TOEIC | `action_question`, `submission_skill_id` |
| 22 | `submit_test_skill` | Submit test | All | `submit_by`, `submission_skill_id` |
| 23 | `system_submit_test_skill` | Submit test | All | `submit_by`, `submission_skill_id` |
| 24 | `tracking_cheating` | Test Room | All | `anti_cheating_type`, `time_of_cheating` |

> `user_id` và `batch_id` là dedicated columns. Các dynamic properties khác nằm trong `properties` JSONB — schema DB không thay đổi khi thêm event mới.

---

## Plan theo Phase

---

### Phase 1: Kafka Foundation

#### 1.1 Kafka Connection Factory

**File mới**: `pkg/kafka/kafka.go`

```go
package kafka

import (
    "crypto/tls"
    "github.com/segmentio/kafka-go"
    "github.com/segmentio/kafka-go/sasl/plain"
)

// NewDialer creates a kafka.Dialer with optional TLS/SASL config
func NewDialer(cfg *KafkaConfig) *kafka.Dialer

// EnsureTopics creates topics if they don't exist (for dev/test environments)
func EnsureTopics(ctx context.Context, brokers []string, dialer *kafka.Dialer, topics []TopicConfig) error
```

#### 1.2 Kafka Producer

**File mới**: `pkg/kafka/producer.go`

```go
type Producer struct {
    writer *kafka.Writer
    logger *zap.Logger
}

func NewProducer(cfg *KafkaConfig, logger *zap.Logger) *Producer

// Produce sends a single message to Kafka
func (p *Producer) Produce(ctx context.Context, topic string, key []byte, value []byte) error

// ProduceBatch sends multiple messages to Kafka
func (p *Producer) ProduceBatch(ctx context.Context, messages []kafka.Message) error

// Close flushes pending messages and closes the writer
func (p *Producer) Close() error
```

Config cho Writer:
- `RequiredAcks = kafka.RequireAll` (-1) — đợi tất cả ISR ack
- `BatchTimeout = 10ms` — low latency cho async produce
- `Async = false` — synchronous để đảm bảo delivery trước khi trả 200

#### 1.3 Kafka Consumer

**File mới**: `pkg/kafka/consumer.go`

```go
type Consumer struct {
    reader    *kafka.Reader
    logger    *zap.Logger
    batchSize int
    flushInterval time.Duration
}

func NewConsumer(cfg *KafkaConfig, logger *zap.Logger) *Consumer

// Start begins consuming messages, calling handler for each batch
func (c *Consumer) Start(ctx context.Context, handler BatchHandler) error

// Close commits offsets and closes the reader
func (c *Consumer) Close() error

// BatchHandler processes a batch of messages
type BatchHandler func(ctx context.Context, messages []kafka.Message) error
```

Consumer config:
- `GroupID` = config value (e.g., `event-tracking-consumer`)
- `MinBytes` = 1, `MaxBytes` = 10MB
- `CommitInterval` = 0 (manual commit)
- `StartOffset` = `kafka.LastOffset`

#### 1.4 Config

**File sửa**: `config/config.go`

```go
type KafkaConfig struct {
    Brokers            []string      // KAFKA_BROKERS (comma-separated)
    Topic              string        // KAFKA_TOPIC = "event-tracking.events"
    DLQTopic           string        // KAFKA_DLQ_TOPIC = "event-tracking.events.dlq"
    ConsumerGroup      string        // KAFKA_CONSUMER_GROUP = "event-tracking-consumer"
    FetchBatchSize     int           // KAFKA_FETCH_BATCH_SIZE = 1500
    FlushIntervalSecs  int           // KAFKA_FLUSH_INTERVAL_SECONDS = 5
    MaxRetries         int           // KAFKA_MAX_RETRIES = 3
    RequiredAcks       int           // KAFKA_REQUIRED_ACKS = -1

    // TLS/SASL (optional, for production)
    TLSEnabled         bool          // KAFKA_TLS_ENABLED
    SASLUsername       string        // KAFKA_SASL_USERNAME
    SASLPassword       string        // KAFKA_SASL_PASSWORD
}
```

**File sửa**: `.env.example`

```env
# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=event-tracking.events
KAFKA_DLQ_TOPIC=event-tracking.events.dlq
KAFKA_CONSUMER_GROUP=event-tracking-consumer
KAFKA_FETCH_BATCH_SIZE=1500
KAFKA_FLUSH_INTERVAL_SECONDS=5
KAFKA_MAX_RETRIES=3
KAFKA_REQUIRED_ACKS=-1
```

**File sửa**: `go.mod` — thêm `github.com/segmentio/kafka-go`

---

### Phase 2: Database & Models (giữ nguyên plan cũ)

#### 2.1 Migration

**File mới**: `migrations/001_create_tracking_events_table.sql`

```sql
CREATE TABLE IF NOT EXISTS tracking_events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event         VARCHAR(100) NOT NULL,
    screen        VARCHAR(100) NOT NULL,
    user_id       BIGINT       NOT NULL,
    batch_id      BIGINT,                          -- nullable, một số event không có batch
    properties    JSONB        DEFAULT '{}',        -- chỉ chứa dynamic fields
    meta_data     JSONB        DEFAULT '{}',
    occurred_at      TIMESTAMPTZ  NOT NULL,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tracking_events_event ON tracking_events (event);
CREATE INDEX idx_tracking_events_screen ON tracking_events (screen);
CREATE INDEX idx_tracking_events_user_id ON tracking_events (user_id);
CREATE INDEX idx_tracking_events_batch_id ON tracking_events (batch_id);
CREATE INDEX idx_tracking_events_occurred_at ON tracking_events (occurred_at);
```

#### 2.2 GORM Model

**File mới**: `internal/models/tracking_event.go`

```go
type TrackingEvent struct {
    models.BaseModel
    Event      string         `gorm:"column:event;type:varchar(100);not null" json:"event"`
    Screen     string         `gorm:"column:screen;type:varchar(100);not null" json:"screen"`
    UserID     int64          `gorm:"column:user_id;not null" json:"user_id"`
    BatchID    *int64         `gorm:"column:batch_id" json:"batch_id,omitempty"`  // nullable
    Properties common.JSONMap `gorm:"column:properties;type:jsonb;default:'{}'" json:"properties"`
    MetaData   common.JSONMap `gorm:"column:meta_data;type:jsonb;default:'{}'" json:"meta_data"`
    OccurredAt    time.Time      `gorm:"column:occurred_at;type:timestamptz;not null" json:"occurred_at"`
    CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (TrackingEvent) TableName() string {
    return "tracking_events"
}
```

#### 2.3 DTO

**File mới**: `internal/dto/tracking_event_dto.go`

```go
type CreateTrackingEventRequest struct {
    Event      string                 `json:"event" binding:"required"`
    Screen     string                 `json:"screen" binding:"required"`
    UserID     int64                  `json:"user_id" binding:"required"`
    BatchID    *int64                 `json:"batch_id,omitempty"`
    Properties map[string]interface{} `json:"properties,omitempty"`    // chỉ dynamic fields
    OccurredAt    *string                `json:"occurred_at,omitempty"`      // ISO 8601, default = now()
}

type CreateBatchTrackingEventRequest struct {
    Events []CreateTrackingEventRequest `json:"events" binding:"required,min=1,max=100,dive"`
}
```

#### 2.4 Repository

**File mới**: `internal/repository/tracking_event_repository.go`

```go
type TrackingEventRepository struct {
    db *gorm.DB
}

func (r *TrackingEventRepository) BulkInsert(ctx context.Context, events []models.TrackingEvent) error {
    return r.db.WithContext(ctx).CreateInBatches(events, 500).Error
}

func (r *TrackingEventRepository) GetByUserID(ctx context.Context, userID int64) ([]models.TrackingEvent, error)
func (r *TrackingEventRepository) GetByBatchID(ctx context.Context, batchID int64) ([]models.TrackingEvent, error)
func (r *TrackingEventRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error)
```

---

### Phase 3: Producer + API Handler

#### 3.1 Event Producer Service

**File mới**: `internal/service/event_producer.go`

```go
type EventProducer struct {
    producer *kafka.Producer
    topic    string
    logger   *zap.Logger
}

func NewEventProducer(producer *kafka.Producer, topic string, logger *zap.Logger) *EventProducer

// Publish serializes DTO to JSON and produces to Kafka
// Partition key: user_id
func (p *EventProducer) Publish(ctx context.Context, event *dto.CreateTrackingEventRequest, metadata map[string]interface{}) error

// PublishBatch publishes multiple events
func (p *EventProducer) PublishBatch(ctx context.Context, events []dto.CreateTrackingEventRequest, metadata map[string]interface{}) error
```

Partition key logic:
```go
func partitionKey(userID int64) string {
    return fmt.Sprintf("user_%d", userID)
}
```

#### 3.2 Handler

**File mới**: `internal/handlers/tracking_event_handler.go`

```go
type TrackingEventHandler struct {
    producer *service.EventProducer
    logger   *zap.Logger
}

// POST /api/v1/events — single event
func (h *TrackingEventHandler) Create(c *gin.Context)

// POST /api/v1/events/batch — batch events (max 100)
func (h *TrackingEventHandler) CreateBatch(c *gin.Context)
```

Luồng xử lý:
1. Validate request (Gin binding)
2. Enrich metadata: gom IP, User-Agent, Device từ request headers vào `meta_data`
3. Produce to Kafka
4. Trả 200 OK

#### 3.3 Routes

**File sửa**: `internal/httpserver/http_start.go`

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

#### 3.4 Main init

**File sửa**: `main.go`

```go
// Init Kafka producer
kafkaProducer := kafka.NewProducer(cfg.Kafka, logger)
defer kafkaProducer.Close()

eventProducer := service.NewEventProducer(kafkaProducer, cfg.Kafka.Topic, logger)

// Pass to HTTP server
```

---

### Phase 4: Consumer + Batch Insert

#### 4.1 Event Consumer Service

**File mới**: `internal/service/event_consumer.go`

```go
type EventConsumer struct {
    consumer    *kafka.Consumer
    dlqProducer *kafka.Producer
    repo        *repository.TrackingEventRepository
    logger      *zap.Logger
    cfg         *config.KafkaConfig
}

func NewEventConsumer(
    consumer *kafka.Consumer,
    dlqProducer *kafka.Producer,
    repo *repository.TrackingEventRepository,
    logger *zap.Logger,
    cfg *config.KafkaConfig,
) *EventConsumer

// Start begins consuming and processing events
func (c *EventConsumer) Start(ctx context.Context) error {
    // Consumer.Start() with batch handler:
    //   1. Deserialize JSON → []dto
    //   2. Transform DTO → []models.TrackingEvent
    //   3. repo.BulkInsert()
    //   4. Nếu lỗi → retry 3 lần (exponential backoff 1s, 2s, 4s)
    //   5. Nếu vẫn lỗi → produce to DLQ topic
    //   6. Commit offsets
}
```

Batch flush logic (trong `pkg/kafka/consumer.go`):
```go
// Accumulate messages until:
//   - Buffer reaches FetchBatchSize (1500), OR
//   - FlushInterval (5s) elapsed since first message in buffer
// Then call BatchHandler with accumulated messages
```

#### 4.2 Graceful shutdown ordering

**File sửa**: `main.go`

```go
// Shutdown sequence:
// 1. Stop HTTP server (stop accepting new requests)
// 2. Close Kafka producer (flush pending messages)
// 3. Close Kafka consumer (commit offsets, finish current batch)
// 4. Close PostgreSQL connection
// 5. Flush observability (traces, logs)
```

---

### Phase 5: Cleanup Redis khỏi event path

**Đánh giá và thực hiện:**

- `internal/scheduler/scheduler.go` — Xoá event processing job (nếu có)
- `main.go` — Đánh giá xoá scheduler init nếu không còn job nào
- `pkg/database/database.go` — Đánh giá xoá Redis init nếu không còn dùng cho mục đích khác

> **Lưu ý**: Redis có thể vẫn cần cho mục đích khác (cache, session, rate limiting). Chỉ xoá phần liên quan đến event buffer.

---

### Phase 6: Observability & Monitoring

#### 6.1 OTel Tracing

- Inject trace context vào Kafka message headers khi produce
- Extract trace context từ Kafka message headers khi consume
- Tạo spans cho: `kafka.produce`, `kafka.consume`, `db.bulk_insert`

#### 6.2 Monitoring Endpoint

**File mới**: `internal/handlers/monitoring_handler.go`

```go
// GET /api/v1/monitoring/consumer-stats (internal API key auth)
// Response:
{
    "consumer_group": "event-tracking-consumer",
    "topic": "event-tracking.events",
    "total_lag": 1234,
    "partitions": [
        {"partition": 0, "current_offset": 1000, "log_end_offset": 1050, "lag": 50},
        ...
    ]
}
```

#### 6.3 Health Check

Cập nhật health check endpoint để kiểm tra:
- PostgreSQL connectivity
- Kafka broker connectivity (Dial test)

---

### Phase 7: Testing

#### Unit Tests

```go
// internal/service/event_producer_test.go
// - Mock kafka.Producer, verify message format & partition key

// internal/service/event_consumer_test.go
// - Mock kafka.Consumer + repository, verify batch processing & DLQ logic

// internal/handlers/tracking_event_handler_test.go
// - Mock EventProducer, test validation & 200 response

// pkg/kafka/producer_test.go
// pkg/kafka/consumer_test.go
```

#### Integration Tests

```go
// Docker compose hoặc testcontainers:
//   - Kafka (confluentinc/cp-kafka)
//   - PostgreSQL
//
// Test flow:
//   1. POST /api/v1/events → verify message in Kafka topic
//   2. Consumer processes → verify record in PostgreSQL
//   3. Simulate DB failure → verify message in DLQ topic
```

---

## Cấu trúc thư mục sau khi implement

```
event-tracking-service/
├── main.go                                    # UPDATE - Kafka init, consumer goroutine, shutdown
├── config/
│   └── config.go                              # UPDATE - KafkaConfig struct
├── migrations/
│   └── 001_create_tracking_events_table.sql   # NEW
├── internal/
│   ├── models/
│   │   ├── base_model.go                      # (existing)
│   │   └── tracking_event.go                  # NEW
│   ├── dto/
│   │   └── tracking_event_dto.go              # NEW
│   ├── repository/
│   │   └── tracking_event_repository.go       # NEW
│   ├── service/
│   │   ├── event_producer.go                  # NEW - Kafka producer wrapper
│   │   └── event_consumer.go                  # NEW - Kafka consumer + batch insert
│   ├── handlers/
│   │   ├── healthcheck_handler.go             # (existing)
│   │   ├── tracking_event_handler.go          # NEW
│   │   └── monitoring_handler.go              # NEW
│   ├── middleware/
│   │   ├── auth.go                            # (existing)
│   │   └── api-key-internal.go                # (existing)
│   ├── httpserver/
│   │   ├── httpserver.go                      # (existing)
│   │   └── http_start.go                      # UPDATE - add event routes
│   └── scheduler/
│       └── scheduler.go                       # UPDATE - remove event processing job
├── pkg/
│   ├── kafka/
│   │   ├── kafka.go                           # NEW - Connection factory, TLS/SASL
│   │   ├── producer.go                        # NEW - Kafka Writer wrapper
│   │   └── consumer.go                        # NEW - Kafka Reader (consumer group)
│   ├── common/                                # (existing)
│   ├── database/                              # (existing)
│   └── observe/                               # (existing)
└── docs/
    ├── plan.md                                # Plan Redis (giữ tham khảo)
    └── plan-kafka.md                          # Plan Kafka (file này)
```

---

## Thứ tự implement (Step by step)

```
Step 1: Kafka Foundation (pkg/kafka/)
   └── kafka.go — Connection factory
   └── producer.go — Writer wrapper
   └── consumer.go — Reader wrapper + batch logic
   └── config/config.go — KafkaConfig
   └── .env.example — Kafka env vars
   └── go.mod — segmentio/kafka-go

Step 2: Database & Models
   └── migrations/001_create_tracking_events_table.sql
   └── internal/models/tracking_event.go
   └── internal/dto/tracking_event_dto.go
   └── internal/repository/tracking_event_repository.go

Step 3: Producer + API Handler
   └── internal/service/event_producer.go
   └── internal/handlers/tracking_event_handler.go
   └── internal/httpserver/http_start.go — add routes
   └── main.go — init Kafka producer

Step 4: Consumer + Batch Insert
   └── internal/service/event_consumer.go
   └── main.go — start consumer goroutine, graceful shutdown

Step 5: Cleanup Redis khỏi event path
   └── internal/scheduler/scheduler.go — remove event job
   └── main.go — evaluate scheduler/Redis removal

Step 6: Observability & Monitoring
   └── OTel tracing (producer/consumer)
   └── internal/handlers/monitoring_handler.go
   └── Health check update

Step 7: Testing
   └── Unit tests
   └── Integration tests
```

---

## So sánh Redis vs Kafka

| Aspect | Redis (plan cũ) | Kafka (plan mới) |
|--------|----------------|-----------------|
| Latency to client | ~1ms | ~5-10ms (vẫn OK cho 200 OK) |
| Processing delay | 0-2 phút (polling) | ~5 giây (continuous consume) |
| Durability | Redis AOF (risk mất data) | Replication (rất bền) |
| Horizontal scaling | Distributed lock | Consumer group tự balance |
| Replay | Không (LPOP destructive) | Có (reset offset) |
| Infrastructure | Đơn giản | Phức tạp hơn (Kafka cluster) |

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
  "properties": {
    "product_line": "IELTS",
    "action_test_direction": "continue"
  },
  "occurred_at": "2026-03-11T10:00:00Z"
}
```

**Response:** `200 OK`
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
      "properties": { "product_line": "IELTS" },
      "occurred_at": "2026-03-11T10:00:00Z"
    },
    {
      "event": "click_agree_exam_regulation",
      "screen": "regulation",
      "user_id": 456,
      "batch_id": 123,
      "properties": { "product_line": "IELTS" },
      "occurred_at": "2026-03-11T10:00:05Z"
    }
  ]
}
```

**Response:** `200 OK`
```json
{
  "message": "2 events accepted",
  "data": null
}
```

### GET /api/v1/monitoring/consumer-stats (Internal)

**Response:**
```json
{
  "message": "ok",
  "data": {
    "consumer_group": "event-tracking-consumer",
    "topic": "event-tracking.events",
    "total_lag": 1234
  }
}
```

---

## Verification

1. **Unit test**: Mock Kafka producer/consumer, test handler logic
2. **Integration test**: Docker compose với Kafka + PostgreSQL, gửi events qua API, verify DB records
3. **Manual test**: `curl POST /api/v1/events` → check Kafka topic → verify PostgreSQL insert
4. **Load test**: Simulate peak exam load, verify consumer throughput và batch insert performance

---

## Lưu ý quan trọng

1. **Hybrid design**: `user_id`, `batch_id` là dedicated columns (query thường xuyên, native index). Các dynamic fields (`action_test_direction`, `part_id`...) vẫn nằm trong `properties` JSONB — không cần migrate khi thêm event mới
2. **At-least-once delivery**: Manual offset commit sau DB write → có thể có duplicate khi consumer restart. Dùng UUID primary key + upsert nếu cần exactly-once
3. **Backward compatible**: Service Go chạy song song với Laravel, không cần migrate data cũ
4. **Performance**: Kafka produce ~5-10ms vẫn OK cho 200 OK pattern
5. **Observability**: Trace context propagation qua Kafka headers giữ distributed tracing liền mạch
6. **Infrastructure**: Cần Kafka cluster (có thể dùng managed service: Confluent Cloud, AWS MSK, Redpanda)
