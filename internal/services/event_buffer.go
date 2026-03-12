package services

import (
	"context"
	"encoding/json"
	"event-tracking-service/config"
	"event-tracking-service/internal/dtos"
	"fmt"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type EventBuffer struct {
	redis     *redis.Client
	logger    *zap.Logger
	queueKey  string
	deadKey   string
	batchSize int
}

func NewEventBuffer(redisClient *redis.Client, cfg *config.Config, logger *zap.Logger) *EventBuffer {
	return &EventBuffer{
		redis:     redisClient,
		logger:    logger,
		queueKey:  cfg.EventBuffer.QueueKey,
		deadKey:   cfg.EventBuffer.DeadLetterKey,
		batchSize: cfg.EventBuffer.BatchSize,
	}
}

func (b *EventBuffer) Push(ctx context.Context, event *dtos.CreateTrackingEventRequest) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if err := b.redis.RPush(ctx, b.queueKey, data).Err(); err != nil {
		return fmt.Errorf("failed to push event to Redis: %w", err)
	}

	return nil
}

func (b *EventBuffer) PushBatch(ctx context.Context, events []dtos.CreateTrackingEventRequest) error {
	pipe := b.redis.Pipeline()
	for i := range events {
		data, err := json.Marshal(&events[i])
		if err != nil {
			return fmt.Errorf("failed to marshal event: %w", err)
		}
		pipe.RPush(ctx, b.queueKey, data)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to push batch events to Redis: %w", err)
	}

	return nil
}

func (b *EventBuffer) PopBatch(ctx context.Context, size int) ([]dtos.CreateTrackingEventRequest, error) {
	pipe := b.redis.Pipeline()
	lrangeCmd := pipe.LRange(ctx, b.queueKey, 0, int64(size-1))
	pipe.LTrim(ctx, b.queueKey, int64(size), -1)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to pop batch from Redis: %w", err)
	}

	rawEvents := lrangeCmd.Val()
	if len(rawEvents) == 0 {
		return nil, nil
	}

	events := make([]dtos.CreateTrackingEventRequest, 0, len(rawEvents))
	for _, raw := range rawEvents {
		var event dtos.CreateTrackingEventRequest
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			b.logger.Error("failed to unmarshal event from Redis", zap.Error(err), zap.String("raw", raw))
			continue
		}
		events = append(events, event)
	}

	return events, nil
}

func (b *EventBuffer) QueueSize(ctx context.Context) (int64, error) {
	return b.redis.LLen(ctx, b.queueKey).Result()
}

func (b *EventBuffer) DeadLetterSize(ctx context.Context) (int64, error) {
	return b.redis.LLen(ctx, b.deadKey).Result()
}

func (b *EventBuffer) PushDeadLetter(ctx context.Context, event *dtos.CreateTrackingEventRequest, reason string) error {
	deadEvent := map[string]interface{}{
		"event":  event,
		"reason": reason,
	}

	data, err := json.Marshal(deadEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal dead letter event: %w", err)
	}

	return b.redis.RPush(ctx, b.deadKey, data).Err()
}

func (b *EventBuffer) BatchSize() int {
	return b.batchSize
}
