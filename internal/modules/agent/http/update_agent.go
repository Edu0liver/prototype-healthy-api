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
// @Failure  400 {object} httputil.ErrorResponse "invalid id or body"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  404 {object} httputil.ErrorResponse "agent not found"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
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
