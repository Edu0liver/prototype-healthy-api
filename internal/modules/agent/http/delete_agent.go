package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Delete handles DELETE /agents/:id.
// @Summary  Delete agent
// @Tags     agents
// @Security BearerAuth
// @Param    id path string true "Agent ID"
// @Success  204
// @Router   /agents/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}
