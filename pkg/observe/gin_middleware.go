package observe

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func GinMiddleware(serviceName string) gin.HandlerFunc {
	otelMiddleware := otelgin.Middleware(serviceName)

	return func(c *gin.Context) {
		skipPaths := []string{
			"/health",
			"/docs",
			"/docs/openapi.yaml",
		}

		for _, path := range skipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		otelMiddleware(c)
	}
}
