package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Update handles PUT /automations/:id.
// @Summary  Update automation
// @Tags     automations
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    id   path string                        true "Automation ID"
// @Param    body body dto.UpdateAutomationRequest   true "Automation"
// @Success  200 {object} dto.AutomationResponse
// @Failure  400 {object} httputil.ErrorResponse "invalid id or body"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  404 {object} httputil.ErrorResponse "automation not found"
// @Failure  409 {object} httputil.ErrorResponse "an active automation already exists for this channel"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /automations/{id} [put]
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
