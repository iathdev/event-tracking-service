package services

import (
	"context"
	"event-tracking-service/config"
	"event-tracking-service/internal/dtos"
	"event-tracking-service/internal/models"
	"event-tracking-service/internal/repositories"
	"event-tracking-service/pkg/common"
	"time"

	"go.uber.org/zap"
)

type EventProcessor struct {
	buffer     *EventBuffer
	repo       repositories.TrackingEventRepository
	logger     *zap.Logger
	maxRetries int
}

func NewEventProcessor(
	buffer *EventBuffer,
	repo repositories.TrackingEventRepository,
	cfg *config.Config,
	logger *zap.Logger,
) *EventProcessor {
	return &EventProcessor{
		buffer:     buffer,
		repo:       repo,
		logger:     logger,
		maxRetries: cfg.EventBuffer.MaxRetries,
	}
}

func (p *EventProcessor) ProcessQueue(ctx context.Context) {
	queueSize, err := p.buffer.QueueSize(ctx)
	if err != nil {
		p.logger.Error("failed to get queue size", zap.Error(err))
		return
	}

	if queueSize == 0 {
		p.logger.Debug("event queue is empty, skipping")
		return
	}

	p.logger.Info("processing event queue", zap.Int64("queue_size", queueSize))

	totalProcessed := 0
	batchSize := p.buffer.BatchSize()

	for {
		if ctx.Err() != nil {
			p.logger.Warn("context cancelled, stopping before next batch", zap.Error(ctx.Err()))
			break
		}

		events, err := p.buffer.PopBatch(ctx, batchSize)
		if err != nil {
			p.logger.Error("failed to pop batch from Redis", zap.Error(err))
			return
		}

		if len(events) == 0 {
			break
		}

		trackingEvents := p.transformEvents(events)

		if err := p.insertWithRetry(ctx, trackingEvents, events); err != nil {
			p.logger.Error("failed to insert events after retries, sending to dead letter",
				zap.Error(err),
				zap.Int("count", len(events)),
			)

			p.sendToDeadLetter(context.Background(), events, err.Error())
		} else {
			totalProcessed += len(trackingEvents)
		}
	}

	p.logger.Info("event queue processing completed", zap.Int("total_processed", totalProcessed))
}

func (p *EventProcessor) transformEvents(dtoEvents []dtos.CreateTrackingEventRequest) []models.TrackingEvent {
	trackingEvents := make([]models.TrackingEvent, 0, len(dtoEvents))

	for _, dto := range dtoEvents {
		occurredAt := time.Now()
		if dto.OccurredAt != nil {
			if parsed, err := time.Parse(time.RFC3339, *dto.OccurredAt); err == nil {
				occurredAt = parsed
			}
		}

		properties := common.JSONMap{}
		for k, v := range dto.Properties {
			properties[k] = v
		}

		metaData := common.JSONMap{}
		for k, v := range dto.MetaData {
			metaData[k] = v
		}

		event := models.TrackingEvent{
			Event:            dto.Event,
			Screen:           dto.Screen,
			UserID:           dto.UserID,
			BatchID:          dto.BatchID,
			Properties:       properties,
			MetaData:         metaData,
			OccurredAt:       occurredAt,
		}

		trackingEvents = append(trackingEvents, event)
	}

	return trackingEvents
}

func (p *EventProcessor) insertWithRetry(ctx context.Context, events []models.TrackingEvent, _ []dtos.CreateTrackingEventRequest) error {
	var lastErr error
	for attempt := 1; attempt <= p.maxRetries; attempt++ {
		if err := p.repo.BulkInsert(ctx, events); err != nil {
			lastErr = err
			p.logger.Warn("bulk insert failed, retrying",
				zap.Int("attempt", attempt),
				zap.Int("max_retries", p.maxRetries),
				zap.Error(err),
			)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}
		return nil
	}
	return lastErr
}

func (p *EventProcessor) sendToDeadLetter(ctx context.Context, events []dtos.CreateTrackingEventRequest, reason string) {
	for i := range events {
		if err := p.buffer.PushDeadLetter(ctx, &events[i], reason); err != nil {
			p.logger.Error("failed to push event to dead letter",
				zap.Error(err),
				zap.String("event", events[i].Event),
			)
		}
	}
}
