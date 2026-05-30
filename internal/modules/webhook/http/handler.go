// Package http exposes the Evolution webhook receiver.
package http

import (
	"io"
	"net/http"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/webhook/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// maxBodyBytes caps a webhook body.
const maxBodyBytes = 5 << 20 // 5 MiB

// Handler receives provider webhooks.
type Handler struct {
	svc   *service.Service
	token string
	log   *zap.Logger
}

// NewHandler builds the handler.
func NewHandler(svc *service.Service, cfg *config.Config, log *zap.Logger) *Handler {
	return &Handler{svc: svc, token: cfg.Evolution.WebhookToken, log: log}
}

// Evolution handles POST /webhooks/evolution. It validates the shared token,
// then processes the event. It always returns 200 quickly so the provider does
// not retry on transient internal errors (those are logged + audited).
// @Summary  Evolution webhook receiver
// @Tags     webhooks
// @Accept   json
// @Produce  json
// @Param    Authorization header string false "Bearer <webhook-token>"
// @Success  200 {object} map[string]bool
// @Router   /webhooks/evolution [post]
func (h *Handler) Evolution(c *gin.Context) {
	if h.token != "" {
		if c.GetHeader("authorization") != "Bearer "+h.token {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyBytes))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad body"})
		return
	}
	if err := h.svc.Process(c.Request.Context(), body); err != nil {
		h.log.Error("webhook processing error", zap.Error(err))
		// Still 200: the event is audited; provider retries add noise.
	}
	c.JSON(http.StatusOK, gin.H{"received": true})
}
