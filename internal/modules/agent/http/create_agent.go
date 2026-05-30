package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Create handles POST /agents.
// @Summary  Create agent
// @Tags     agents
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.CreateAgentRequest true "Agent"
// @Success  201 {object} dto.AgentResponse
// @Router   /agents [post]
func (h *Handler) Create(c *gin.Context) {
	var in dto.CreateAgentRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	a, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, agentResponse(a))
}
