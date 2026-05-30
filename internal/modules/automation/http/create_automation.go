package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Create handles POST /automations.
func (h *Handler) Create(c *gin.Context) {
	var in dto.CreateAutomationRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	a, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, automationResponse(a))
}
