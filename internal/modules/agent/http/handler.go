// Package http exposes the agent module's Gin handlers.
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/dtos"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/services"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler serves agent CRUD endpoints.
type Handler struct {
	svc *services.Service
}

// NewHandler builds the handler.
func NewHandler(svc *services.Service) *Handler { return &Handler{svc: svc} }

// Create handles POST /agents.
func (h *Handler) Create(c *gin.Context) {
	var in dtos.CreateAgentRequest
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

// Update handles PUT /agents/:id.
func (h *Handler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in dtos.UpdateAgentRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	a, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, agentResponse(a))
}

// Get handles GET /agents/:id.
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

// List handles GET /agents.
func (h *Handler) List(c *gin.Context) {
	agents, err := h.svc.List(c.Request.Context())
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dtos.AgentResponse, len(agents))
	for i := range agents {
		out[i] = agentResponse(&agents[i])
	}
	httputil.OK(c, gin.H{"agents": out})
}

// Delete handles DELETE /agents/:id.
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

func agentResponse(a *models.Agent) dtos.AgentResponse {
	return dtos.AgentResponse{
		ID: a.ID.String(), Name: a.Name, SystemPrompt: a.SystemPrompt, Model: a.Model,
		Temperature: a.Temperature, MaxOutputTokens: a.MaxOutputTokens,
		HandoverEnabled: a.HandoverEnabled, HandoverKeywords: []string(a.HandoverKeywords), Status: a.Status,
	}
}
