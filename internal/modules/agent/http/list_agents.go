package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// List handles GET /agents.
// @Summary  List agents
// @Tags     agents
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} map[string][]dto.AgentResponse
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /agents [get]
func (h *Handler) List(c *gin.Context) {
	agents, err := h.svc.List(c.Request.Context())
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dto.AgentResponse, len(agents))
	for i := range agents {
		out[i] = agentResponse(&agents[i])
	}
	httputil.OK(c, gin.H{"agents": out})
}
