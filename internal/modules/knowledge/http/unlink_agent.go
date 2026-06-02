package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// UnlinkAgent handles DELETE /agents/:id/knowledge-bases/:kbId.
// @Summary  Unlink knowledge base from agent
// @Tags     knowledge
// @Security BearerAuth
// @Param    id   path string true "Agent ID"
// @Param    kbId path string true "Knowledge base ID"
// @Success  204
// @Failure  400 {object} httputil.ErrorResponse "invalid id"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  404 {object} httputil.ErrorResponse "agent or knowledge base not found"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /agents/{id}/knowledge-bases/{kbId} [delete]
func (h *Handler) UnlinkAgent(c *gin.Context) {
	agentID, ok := parseID(c, "id")
	if !ok {
		return
	}
	kbID, ok := parseID(c, "kbId")
	if !ok {
		return
	}
	if err := h.svc.UnlinkAgent(c.Request.Context(), agentID, kbID); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}
