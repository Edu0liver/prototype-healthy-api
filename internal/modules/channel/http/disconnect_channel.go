package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Disconnect handles DELETE /channels/:id.
// @Summary  Disconnect channel
// @Tags     channels
// @Security BearerAuth
// @Param    id path string true "Channel ID"
// @Success  204
// @Router   /channels/{id} [delete]
func (h *Handler) Disconnect(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Disconnect(c.Request.Context(), id); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}
