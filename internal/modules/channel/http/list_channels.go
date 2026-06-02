package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// List handles GET /channels.
// @Summary  List channels
// @Tags     channels
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} map[string][]dto.ChannelResponse
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /channels [get]
func (h *Handler) List(c *gin.Context) {
	channels, err := h.svc.List(c.Request.Context())
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dto.ChannelResponse, len(channels))
	for i := range channels {
		out[i] = channelResponse(&channels[i])
	}
	httputil.OK(c, gin.H{"channels": out})
}
