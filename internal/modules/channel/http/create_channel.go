package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Create handles POST /channels.
// @Summary  Create channel
// @Tags     channels
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.CreateChannelRequest true "Channel"
// @Success  201 {object} dto.ChannelResponse
// @Failure  400 {object} httputil.ErrorResponse "unsupported channel type or invalid body"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /channels [post]
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
