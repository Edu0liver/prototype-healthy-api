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
// @Failure  400 {object} httputil.ErrorResponse "invalid id"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  404 {object} httputil.ErrorResponse "knowledge base not found"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
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
