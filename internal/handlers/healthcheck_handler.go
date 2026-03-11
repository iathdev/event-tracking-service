package handlers

import (
	"event-tracking-service/pkg/common"

	"github.com/gin-gonic/gin"
)

type HealthCheckHandler interface {
	Health(ctx *gin.Context)
}

type healthCheckHandler struct {
}

func NewHealthCheckHandler() HealthCheckHandler {
	return &healthCheckHandler{}
}

func (h *healthCheckHandler) Health(ctx *gin.Context) {
	common.WriteSuccessResponse(ctx, "OK", nil)
}
