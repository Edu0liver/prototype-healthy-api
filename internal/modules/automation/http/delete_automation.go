package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Delete handles DELETE /automations/:id.
// @Summary  Delete automation
// @Tags     automations
// @Security BearerAuth
// @Param    id path string true "Automation ID"
// @Success  204
// @Router   /automations/{id} [delete]
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
