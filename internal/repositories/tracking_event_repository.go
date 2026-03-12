package repositories

import (
	"context"
	"event-tracking-service/internal/models"

	"gorm.io/gorm"
)

type TrackingEventRepository interface {
	BulkInsert(ctx context.Context, events []models.TrackingEvent) error
	GetByUserID(ctx context.Context, userID int64) ([]models.TrackingEvent, error)
	GetByBatchID(ctx context.Context, batchID int64) ([]models.TrackingEvent, error)
	DeleteOlderThan(ctx context.Context, days int) (int64, error)
}

type trackingEventRepository struct {
	db *gorm.DB
}

func NewTrackingEventRepository(db *gorm.DB) TrackingEventRepository {
	return &trackingEventRepository{db: db}
}

func (r *trackingEventRepository) BulkInsert(ctx context.Context, events []models.TrackingEvent) error {
	return r.db.WithContext(ctx).CreateInBatches(events, 1000).Error
}

func (r *trackingEventRepository) GetByUserID(ctx context.Context, userID int64) ([]models.TrackingEvent, error) {
	var events []models.TrackingEvent
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&events).Error
	return events, err
}

func (r *trackingEventRepository) GetByBatchID(ctx context.Context, batchID int64) ([]models.TrackingEvent, error) {
	var events []models.TrackingEvent
	err := r.db.WithContext(ctx).Where("batch_id = ?", batchID).Find(&events).Error
	return events, err
}

func (r *trackingEventRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("created_at < NOW() - INTERVAL '1 day' * ?", days).
		Delete(&models.TrackingEvent{})
	return result.RowsAffected, result.Error
}
