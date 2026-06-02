package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Create handles POST /automations.
// @Summary  Create automation
// @Tags     automations
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.CreateAutomationRequest true "Automation"
// @Success  201 {object} dto.AutomationResponse
// @Failure  400 {object} httputil.ErrorResponse "invalid body, or channel/agent not found in tenant"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  409 {object} httputil.ErrorResponse "an active automation already exists for this channel"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /automations [post]
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
