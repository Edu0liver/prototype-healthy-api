// Package http exposes the channel module's Gin handlers (one file per use case).
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/service"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler serves channel endpoints.
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

func channelResponse(ch *models.Channel) dto.ChannelResponse {
	var agentID *string
	if ch.ActiveAgentID != nil {
		s := ch.ActiveAgentID.String()
		agentID = &s
	}
	return dto.ChannelResponse{
		ID: ch.ID.String(), Type: ch.Type, Name: ch.Name, Status: ch.Status,
		ExternalAccountID: ch.ExternalAccountID, InstanceName: ch.EvolutionInstanceName, ActiveAgentID: agentID,
	}
}
