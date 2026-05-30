package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Update handles PUT /agents/:id.
// @Summary  Update agent
// @Tags     agents
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    id   path string                 true "Agent ID"
// @Param    body body dto.UpdateAgentRequest true "Agent"
// @Success  200 {object} dto.AgentResponse
// @Router   /agents/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in dto.UpdateAgentRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	a, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, agentResponse(a))
}
