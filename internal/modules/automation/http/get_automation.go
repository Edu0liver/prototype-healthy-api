package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Get handles GET /automations/:id.
// @Summary  Get automation
// @Tags     automations
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Automation ID"
// @Success  200 {object} dto.AutomationResponse
// @Router   /automations/{id} [get]
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
	httputil.OK(c, automationResponse(a))
}

// List handles GET /automations.
// @Summary  List automations
// @Tags     automations
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} map[string][]dto.AutomationResponse
// @Router   /automations [get]
func (h *Handler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dto.AutomationResponse, len(items))
	for i := range items {
		out[i] = automationResponse(&items[i])
	}
	httputil.OK(c, gin.H{"automations": out})
}
