// Package http exposes the conversation module's panel handlers (RF-LOG).
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler serves conversation history/listing endpoints.
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

func conversationResponse(c *models.Conversation) dto.ConversationResponse {
	r := dto.ConversationResponse{
		ID: c.ID.String(), ChannelID: c.ChannelID.String(), ContactID: c.ContactID.String(),
		State: c.State, LastMessageAt: c.LastMessageAt, OpenedAt: c.OpenedAt, ClosedAt: c.ClosedAt,
	}
	if c.AgentID != nil {
		s := c.AgentID.String()
		r.AgentID = &s
	}
	if c.AssignedUserID != nil {
		s := c.AssignedUserID.String()
		r.AssignedUserID = &s
	}
	return r
}

func messageResponse(m *models.Message) dto.MessageResponse {
	return dto.MessageResponse{
		ID: m.ID.String(), Direction: m.Direction, SenderType: m.SenderType, Content: m.Content,
		Media: m.Media, ExternalMessageID: m.ExternalMessageID, Status: m.Status, CreatedAt: m.CreatedAt,
	}
}
