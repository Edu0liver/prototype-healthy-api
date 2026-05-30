package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Return handles POST /conversations/:id/handover/return.
// @Summary  Return conversation to AI
// @Tags     handover
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Conversation ID"
// @Success  200 {object} map[string]string
// @Router   /conversations/{id}/handover/return [post]
func (h *Handler) Return(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Return(c.Request.Context(), id); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, gin.H{"state": "ai"})
}
