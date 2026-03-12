package models

import (
	"event-tracking-service/pkg/common"
	"time"
)

type TrackingEvent struct {
	BaseModel
	Event            string         `gorm:"column:event;type:varchar(100);not null" json:"event"`
	Screen           string         `gorm:"column:screen;type:varchar(100);not null" json:"screen"`
	UserID           int64          `gorm:"column:user_id;not null" json:"user_id"`
	BatchID          *int64         `gorm:"column:batch_id" json:"batch_id,omitempty"`
	Properties       common.JSONMap `gorm:"column:properties;type:jsonb;default:'{}'" json:"properties"`
	MetaData         common.JSONMap `gorm:"column:meta_data;type:jsonb;default:'{}'" json:"meta_data"`
	OccurredAt       time.Time      `gorm:"primaryKey;column:occurred_at;type:timestamptz;not null" json:"occurred_at"`
	CreatedAt        time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (TrackingEvent) TableName() string {
	return "tracking_events"
}
