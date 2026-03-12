package handlers

import (
	"event-tracking-service/internal/services"
	"event-tracking-service/pkg/common"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type MonitoringHandler interface {
	QueueStats(ctx *gin.Context)
}

type monitoringHandler struct {
	buffer *services.EventBuffer
	logger *zap.Logger
}

func NewMonitoringHandler(buffer *services.EventBuffer, logger *zap.Logger) MonitoringHandler {
	return &monitoringHandler{
		buffer: buffer,
		logger: logger,
	}
}

func (h *monitoringHandler) QueueStats(ctx *gin.Context) {
	queueSize, err := h.buffer.QueueSize(ctx.Request.Context())
	if err != nil {
		h.logger.Error("failed to get queue size", zap.Error(err))
		queueSize = -1
	}

	deadLetterSize, err := h.buffer.DeadLetterSize(ctx.Request.Context())
	if err != nil {
		h.logger.Error("failed to get dead letter size", zap.Error(err))
		deadLetterSize = -1
	}

	common.WriteSuccessResponse(ctx, "ok", gin.H{
		"queue_size":       queueSize,
		"dead_letter_size": deadLetterSize,
	})
}
