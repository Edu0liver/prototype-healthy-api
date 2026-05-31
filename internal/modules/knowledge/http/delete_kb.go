package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// DeleteKB handles DELETE /knowledge-bases/:id.
// @Summary  Delete knowledge base
// @Tags     knowledge
// @Security BearerAuth
// @Param    id path string true "Knowledge base ID"
// @Success  204
// @Router   /knowledge-bases/{id} [delete]
func (h *Handler) DeleteKB(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteKB(c.Request.Context(), id); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}
