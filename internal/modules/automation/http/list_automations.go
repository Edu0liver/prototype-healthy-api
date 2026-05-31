package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

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
