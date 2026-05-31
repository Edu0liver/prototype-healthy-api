package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// ListAgentKBs handles GET /agents/:id/knowledge-bases, returning the ids of
// the knowledge bases currently linked to the agent (RF-AG-02).
// @Summary  List knowledge bases linked to an agent
// @Tags     knowledge
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Agent ID"
// @Success  200 {object} map[string][]string
// @Router   /agents/{id}/knowledge-bases [get]
func (h *Handler) ListAgentKBs(c *gin.Context) {
	agentID, ok := parseID(c, "id")
	if !ok {
		return
	}
	ids, err := h.svc.KBIDsForAgent(c.Request.Context(), agentID)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = id.String()
	}
	httputil.OK(c, gin.H{"knowledge_base_ids": out})
}
