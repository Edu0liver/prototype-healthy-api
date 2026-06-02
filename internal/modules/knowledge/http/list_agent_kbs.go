package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// ListAgentKBs handles GET /agents/:id/knowledge-bases.
// @Summary  List knowledge bases linked to an agent
// @Tags     knowledge
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Agent ID"
// @Success  200 {object} map[string][]dto.KBResponse
// @Failure  400 {object} httputil.ErrorResponse "invalid id"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  404 {object} httputil.ErrorResponse "agent not found"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /agents/{id}/knowledge-bases [get]
func (h *Handler) ListAgentKBs(c *gin.Context) {
	agentID, ok := parseID(c, "id")
	if !ok {
		return
	}
	kbs, err := h.svc.KBsForAgent(c.Request.Context(), agentID)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dto.KBResponse, len(kbs))
	for i := range kbs {
		out[i] = kbResponse(&kbs[i])
	}
	httputil.OK(c, gin.H{"knowledge_bases": out})
}
