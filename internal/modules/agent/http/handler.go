// Package http exposes the agent module's Gin handlers (one file per use case).
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/service"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler serves agent CRUD endpoints.
type Handler struct {
	svc *service.Service
}

// NewHandler builds the handler.
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Fail(c, httputil.BadRequest("invalid id"))
		return uuid.Nil, false
	}
	return id, true
}

func agentResponse(a *models.Agent) dto.AgentResponse {
	return dto.AgentResponse{
		ID: a.ID.String(), Name: a.Name, SystemPrompt: a.SystemPrompt, Model: a.Model,
		Temperature: a.Temperature, MaxOutputTokens: a.MaxOutputTokens,
		HandoverEnabled: a.HandoverEnabled, HandoverKeywords: []string(a.HandoverKeywords), Status: a.Status,
	}
}
