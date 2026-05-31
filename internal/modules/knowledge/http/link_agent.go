package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// LinkAgent handles POST /agents/:id/knowledge-bases/:kbId.
// @Summary  Link knowledge base to agent
// @Tags     knowledge
// @Security BearerAuth
// @Param    id   path string true "Agent ID"
// @Param    kbId path string true "Knowledge base ID"
// @Success  204
// @Router   /agents/{id}/knowledge-bases/{kbId} [post]
func (h *Handler) LinkAgent(c *gin.Context) {
	agentID, ok := parseID(c, "id")
	if !ok {
		return
	}
	kbID, ok := parseID(c, "kbId")
	if !ok {
		return
	}
	if err := h.svc.LinkAgent(c.Request.Context(), agentID, kbID); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}
