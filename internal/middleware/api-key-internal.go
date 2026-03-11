package middleware

import (
	"event-tracking-service/pkg/common"

	"github.com/gin-gonic/gin"
)

func RequiredApiKeyInternal() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKeyHeader := c.GetHeader("x-api-key")
		apiKey := common.GetEnv("API_KEY_INTERNAL", "")
		if apiKey == "" || apiKeyHeader == "" {
			common.WriteErrorUnauthorizedResponse(c, "API KEY invalid", nil)
			c.Abort()
			return
		}

		if apiKey != apiKeyHeader {
			common.WriteErrorUnauthorizedResponse(c, "API KEY invalid", nil)
			c.Abort()
			return
		}

		c.Next()
	}
}
