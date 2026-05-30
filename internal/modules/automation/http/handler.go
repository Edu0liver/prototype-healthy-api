// Package http exposes the automation module's Gin handlers (one file per use case).
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/service"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler serves automation endpoints.
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

func automationResponse(a *models.Automation) dto.AutomationResponse {
	return dto.AutomationResponse{
		ID: a.ID.String(), ChannelID: a.ChannelID.String(), AgentID: a.AgentID.String(),
		IsActive: a.IsActive, FallbackMessage: a.FallbackMessage, DebounceSeconds: a.DebounceSeconds,
	}
}
