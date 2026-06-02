// Package http exposes the Evolution webhook receiver.
package http

import (
	"crypto/subtle"
	"io"
	"net/http"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/webhook/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// maxBodyBytes caps a webhook body.
const maxBodyBytes = 5 << 20 // 5 MiB

// AckResponse is the webhook acknowledgement payload.
type AckResponse struct {
	Received bool `json:"received"`
}

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
// @Success  200 {object} AckResponse
// @Failure  400 {object} httputil.ErrorResponse "bad request body"
// @Failure  401 {object} httputil.ErrorResponse "invalid webhook token"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /webhooks/evolution [post]
func (h *Handler) Evolution(c *gin.Context) {
	h.log.Debug("webhook: request received", zap.String("remote_addr", c.Request.RemoteAddr))
	if h.token != "" {
		want := "Bearer " + h.token
		got := c.GetHeader("authorization")
		if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
			h.log.Warn("webhook: auth failed", zap.String("remote_addr", c.Request.RemoteAddr))
			c.AbortWithStatusJSON(http.StatusUnauthorized, httputil.ErrorResponse{Error: "unauthorized", Message: "invalid webhook token"})
			return
		}
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBodyBytes))
	if err != nil {
		h.log.Error("webhook: read body failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, httputil.ErrorResponse{Error: "bad_request", Message: "bad body"})
		return
	}
	h.log.Debug("webhook: raw body", zap.Int("bytes", len(body)))
	h.log.Debug("webhook: calling svc.Process")
	if err := h.svc.Process(c.Request.Context(), body); err != nil {
		h.log.Error("webhook: svc.Process error", zap.Error(err))
		// Still 200: the event is audited; provider retries add noise.
	} else {
		h.log.Debug("webhook: svc.Process completed ok")
	}
	c.JSON(http.StatusOK, AckResponse{Received: true})
}
