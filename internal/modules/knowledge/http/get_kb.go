package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// GetKB handles GET /knowledge-bases/:id.
// @Summary  Get knowledge base
// @Tags     knowledge
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Knowledge base ID"
// @Success  200
// @Router   /knowledge-bases/{id} [get]
func (h *Handler) GetKB(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	kb, err := h.svc.GetKB(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, kbResponse(kb))
}
