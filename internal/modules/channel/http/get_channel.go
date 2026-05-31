package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Get handles GET /channels/:id.
// @Summary  Get channel
// @Tags     channels
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Channel ID"
// @Success  200
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
