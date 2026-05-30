package http

import "github.com/gin-gonic/gin"

// RegisterRoutes mounts the public webhook endpoint (auth via shared token, not JWT).
func RegisterRoutes(e *gin.Engine, h *Handler) {
	e.POST("/webhooks/evolution", h.Evolution)
}
