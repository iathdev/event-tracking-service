package common

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type coreResponse struct {
	Success  bool        `json:"success"`
	Message  string      `json:"message"`
	Error    interface{} `json:"error,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	MetaData interface{} `json:"metadata,omitempty"`
}

func WriteSuccessResponse(ctx *gin.Context, message string, data interface{}) {
	res := &coreResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	ctx.JSON(http.StatusOK, res)
}

func WriteErrorResponse(c *gin.Context, message string, err interface{}) {
	res := &coreResponse{
		Success: false,
		Message: message,
		Error:   err,
	}

	c.JSON(http.StatusBadRequest, res)
}

func WriteErrorResponseWithCode(c *gin.Context, message string, err interface{}, code int) {
	res := &coreResponse{
		Success: false,
		Message: message,
		Error:   err,
	}

	c.JSON(code, res)
}

func WriteErrorUnauthorizedResponse(c *gin.Context, message string, err interface{}) {
	res := &coreResponse{
		Success: false,
		Message: message,
		Error:   err,
	}

	c.JSON(http.StatusUnauthorized, res)
}

func WriteErrorValidatorResponse(c *gin.Context, message string, err interface{}) {
	res := &coreResponse{
		Success: false,
		Message: message,
		Error:   err,
	}

	c.JSON(http.StatusUnprocessableEntity, res)
}

func WriteSuccessWithPaginate(ctx *gin.Context, message string, data interface{}, meta interface{}) {
	res := &coreResponse{
		Success:  true,
		Message:  message,
		Data:     data,
		MetaData: meta,
	}

	ctx.JSON(http.StatusOK, res)
}
