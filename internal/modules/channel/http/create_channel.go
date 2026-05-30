package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Create handles POST /channels.
func (h *Handler) Create(c *gin.Context) {
	var in dto.CreateChannelRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	ch, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, channelResponse(ch))
}
