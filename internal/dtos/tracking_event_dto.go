package dtos

type CreateTrackingEventRequest struct {
	Event            string                 `json:"event" binding:"required"`
	Screen           string                 `json:"screen" binding:"required"`
	UserID           int64                  `json:"user_id" binding:"required"`
	BatchID          *int64                 `json:"batch_id,omitempty"`
	Properties       map[string]interface{} `json:"properties,omitempty"`
	MetaData         map[string]interface{} `json:"meta_data,omitempty"`
	OccurredAt       *string                `json:"occurred_at,omitempty"`
}

type CreateBatchTrackingEventRequest struct {
	Events []CreateTrackingEventRequest `json:"events" binding:"required,min=1,max=100,dive"`
}
