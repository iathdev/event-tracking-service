package observe

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func LoggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Skip logging for health and docs endpoints
		skipPaths := []string{"/health", "/docs", "/docs/openapi.yaml"}
		for _, sp := range skipPaths {
			if path == sp {
				c.Next()
				return
			}
		}

		// Read request body
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			requestBody = string(bodyBytes)
			if len(requestBody) > 1000 {
				requestBody = requestBody[:1000] + "..."
			}
		}

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
		}

		// Add trace ID if available
		span := trace.SpanFromContext(c.Request.Context())
		if span.SpanContext().IsValid() {
			fields = append(fields,
				zap.String("trace_id", span.SpanContext().TraceID().String()),
				zap.String("span_id", span.SpanContext().SpanID().String()),
			)
		}

		// Add request body for non-GET requests
		if method != "GET" && requestBody != "" {
			fields = append(fields, zap.String("request_body", requestBody))
		}

		// Log based on status code
		if status >= 500 {
			logger.Error("HTTP Request", fields...)
		} else if status >= 400 {
			logger.Warn("HTTP Request", fields...)
		} else {
			logger.Info("HTTP Request", fields...)
		}
	}
}

func RecoveryMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.Stack("stacktrace"),
				)
				c.AbortWithStatusJSON(500, gin.H{
					"message": "Internal Server Error",
					"error":   err,
				})
			}
		}()
		c.Next()
	}
}
