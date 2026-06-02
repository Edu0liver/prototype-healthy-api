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
	h.log.Debug("webhook: request received", zap.String("remote_addr", c.Request.RemoteAddr))
	if h.token != "" {
		if c.GetHeader("authorization") != "Bearer "+h.token {
			h.log.Warn("webhook: auth failed", zap.String("got", c.GetHeader("authorization")))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyBytes))
	if err != nil {
		h.log.Error("webhook: read body failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad body"})
		return
	}
	h.log.Debug("webhook: raw body", zap.Int("bytes", len(body)), zap.String("body", string(body)))
	h.log.Debug("webhook: calling svc.Process")
	if err := h.svc.Process(c.Request.Context(), body); err != nil {
		h.log.Error("webhook: svc.Process error", zap.Error(err))
		// Still 200: the event is audited; provider retries add noise.
	} else {
		h.log.Debug("webhook: svc.Process completed ok")
	}
	c.JSON(http.StatusOK, gin.H{"received": true})
}
