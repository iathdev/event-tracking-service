package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type DocsHandler struct{}

func NewDocsHandler() *DocsHandler {
	return &DocsHandler{}
}

func (h *DocsHandler) ScalarUI(c *gin.Context) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Event Tracking Service API Documentation</title>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
    <script id="api-reference" data-url="/docs/openapi.yaml"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func (h *DocsHandler) OpenAPISpec(c *gin.Context) {
	content, err := os.ReadFile("docs/api/openapi.yaml")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to read OpenAPI spec"})
		return
	}

	c.Header("Content-Type", "application/x-yaml")
	c.String(http.StatusOK, string(content))
}
