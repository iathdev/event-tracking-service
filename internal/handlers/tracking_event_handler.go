package handlers

import (
	"event-tracking-service/internal/dtos"
	"event-tracking-service/internal/services"
	"event-tracking-service/pkg/common"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TrackingEventHandler interface {
	Create(ctx *gin.Context)
	CreateBatch(ctx *gin.Context)
}

type trackingEventHandler struct {
	buffer          *services.EventBuffer
	logger          *zap.Logger
	sensitiveFields map[string]struct{}
}

func NewTrackingEventHandler(buffer *services.EventBuffer, logger *zap.Logger, sensitiveFields []string) TrackingEventHandler {
	fieldSet := make(map[string]struct{}, len(sensitiveFields))
	for _, f := range sensitiveFields {
		fieldSet[f] = struct{}{}
	}

	return &trackingEventHandler{
		buffer:          buffer,
		logger:          logger,
		sensitiveFields: fieldSet,
	}
}

func (h *trackingEventHandler) Create(ctx *gin.Context) {
	var req dtos.CreateTrackingEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.WriteErrorValidatorResponse(ctx, "Validation failed", err.Error())
		return
	}

	enrichMetaData(ctx, &req)
	h.stripSensitiveFields(&req)

	if err := h.buffer.Push(ctx.Request.Context(), &req); err != nil {
		h.logger.Error("failed to push event to buffer", zap.Error(err))
		common.WriteErrorResponseWithCode(ctx, "Failed to accept event", err.Error(), http.StatusInternalServerError)
		return
	}

	common.WriteSuccessResponse(ctx, "Event accepted", nil)
}

func (h *trackingEventHandler) CreateBatch(ctx *gin.Context) {
	var req dtos.CreateBatchTrackingEventRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.WriteErrorValidatorResponse(ctx, "Validation failed", err.Error())
		return
	}

	for i := range req.Events {
		enrichMetaData(ctx, &req.Events[i])
		h.stripSensitiveFields(&req.Events[i])
	}

	if err := h.buffer.PushBatch(ctx.Request.Context(), req.Events); err != nil {
		h.logger.Error("failed to push batch events to buffer", zap.Error(err))
		common.WriteErrorResponseWithCode(ctx, "Failed to accept events", nil, http.StatusInternalServerError)
		return
	}

	common.WriteSuccessResponse(ctx, fmt.Sprintf("%d events accepted", len(req.Events)), nil)
}

func enrichMetaData(ctx *gin.Context, req *dtos.CreateTrackingEventRequest) {
	if req.MetaData == nil {
		req.MetaData = make(map[string]interface{})
	}

	req.MetaData["ip"] = ctx.ClientIP()
	req.MetaData["user_agent"] = ctx.GetHeader("User-Agent")
}

func (h *trackingEventHandler) stripSensitiveFields(req *dtos.CreateTrackingEventRequest) {
	if len(h.sensitiveFields) == 0 || len(req.Properties) == 0 {
		return
	}

	for key := range req.Properties {
		if _, isSensitive := h.sensitiveFields[key]; isSensitive {
			delete(req.Properties, key)
		}
	}
}
