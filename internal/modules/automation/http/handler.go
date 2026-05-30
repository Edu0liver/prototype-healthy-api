// Package http exposes the automation module's Gin handlers.
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/dtos"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/services"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler serves automation endpoints.
type Handler struct {
	svc *services.Service
}

// NewHandler builds the handler.
func NewHandler(svc *services.Service) *Handler { return &Handler{svc: svc} }

// Create handles POST /automations.
func (h *Handler) Create(c *gin.Context) {
	var in dtos.CreateAutomationRequest
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

// Update handles PUT /automations/:id.
func (h *Handler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in dtos.UpdateAutomationRequest
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

// Get handles GET /automations/:id.
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
func (h *Handler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dtos.AutomationResponse, len(items))
	for i := range items {
		out[i] = automationResponse(&items[i])
	}
	httputil.OK(c, gin.H{"automations": out})
}

// Delete handles DELETE /automations/:id.
func (h *Handler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Fail(c, httputil.BadRequest("invalid id"))
		return uuid.Nil, false
	}
	return id, true
}

func automationResponse(a *models.Automation) dtos.AutomationResponse {
	return dtos.AutomationResponse{
		ID: a.ID.String(), ChannelID: a.ChannelID.String(), AgentID: a.AgentID.String(),
		IsActive: a.IsActive, FallbackMessage: a.FallbackMessage, DebounceSeconds: a.DebounceSeconds,
	}
}
