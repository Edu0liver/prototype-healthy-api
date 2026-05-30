package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Get handles GET /channels/:id.
// @Summary  Get channel
// @Tags     channels
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Channel ID"
// @Success  200 {object} dto.ChannelResponse
// @Router   /channels/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	ch, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, channelResponse(ch))
}

// List handles GET /channels.
// @Summary  List channels
// @Tags     channels
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} map[string][]dto.ChannelResponse
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
