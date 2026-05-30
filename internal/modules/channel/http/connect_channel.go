package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Connect handles POST /channels/:id/connect.
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
