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
// @Router   /channels/{id}/connect [post]
func (h *Handler) Connect(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in dto.ConnectRequest
	_ = c.ShouldBindJSON(&in)
	if in.Method == "" {
		in.Method = "qr"
	}
	res, err := h.svc.Connect(c.Request.Context(), id, in.Method, in.Number)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, res)
}

// ConnectionState handles GET /channels/:id/connection-state.
// @Summary  Get connection state
// @Tags     channels
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Channel ID"
// @Success  200 {object} map[string]string
// @Router   /channels/{id}/connection-state [get]
func (h *Handler) ConnectionState(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	ch, err := h.svc.RefreshState(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, gin.H{"state": ch.Status})
}
