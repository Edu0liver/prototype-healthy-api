package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Update handles PUT /automations/:id.
func (h *Handler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in dto.UpdateAutomationRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	a, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, automationResponse(a))
}
