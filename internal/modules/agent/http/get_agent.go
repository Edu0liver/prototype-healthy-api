package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Get handles GET /agents/:id.
// @Summary  Get agent
// @Tags     agents
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Agent ID"
// @Success  200
// @Router   /agents/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	a, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, agentResponse(a))
}
