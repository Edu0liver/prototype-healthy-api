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
// @Failure  400 {object} httputil.ErrorResponse "invalid request"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
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
