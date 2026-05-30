package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Close handles POST /conversations/:id/handover/close.
// @Summary  Close conversation
// @Tags     handover
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Conversation ID"
// @Success  200 {object} map[string]string
// @Router   /conversations/{id}/handover/close [post]
func (h *Handler) Close(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Close(c.Request.Context(), id); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, gin.H{"state": "closed"})
}
