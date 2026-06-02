package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Connect handles POST /channels/:id/connect.
// @Summary  Connect channel
// @Tags     channels
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    id   path string             true  "Channel ID"
// @Param    body body dto.ConnectRequest false "Connect options (method: qr|pairing)"
// @Success  200 {object} dto.ConnectResponse
// @Failure  400 {object} httputil.ErrorResponse "invalid id or not a whatsapp channel"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  404 {object} httputil.ErrorResponse "channel not found"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /channels/{id}/connect [post]
func (h *Handler) Connect(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in dto.ConnectRequest
	_ = c.ShouldBindJSON(&in)
	res, err := h.svc.Connect(c.Request.Context(), id, in.Number)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, res)
}
