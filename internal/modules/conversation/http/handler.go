// Package http exposes the conversation module's panel handlers (RF-LOG).
package http

import (
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/dtos"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/services"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler serves conversation history/listing endpoints.
type Handler struct {
	svc *services.Service
}

// NewHandler builds the handler.
func NewHandler(svc *services.Service) *Handler { return &Handler{svc: svc} }

// List handles GET /conversations with optional filters (RF-LOG-03).
func (h *Handler) List(c *gin.Context) {
	f := repositories.ConversationFilter{State: c.Query("state")}
	if v := c.Query("channel_id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			f.ChannelID = &id
		}
	}
	if v := c.Query("since"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			f.Since = &t
		}
	}
	items, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dtos.ConversationResponse, len(items))
	for i := range items {
		out[i] = conversationResponse(&items[i])
	}
	httputil.OK(c, gin.H{"conversations": out})
}

// Get handles GET /conversations/:id.
func (h *Handler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	conv, err := h.svc.GetConversation(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, conversationResponse(conv))
}

// Messages handles GET /conversations/:id/messages (RF-LOG-02 full audit).
func (h *Handler) Messages(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if _, err := h.svc.GetConversation(c.Request.Context(), id); err != nil {
		httputil.Fail(c, err)
		return
	}
	msgs, err := h.svc.Messages(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dtos.MessageResponse, len(msgs))
	for i := range msgs {
		out[i] = messageResponse(&msgs[i])
	}
	httputil.OK(c, gin.H{"messages": out})
}

func parseID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httputil.Fail(c, httputil.BadRequest("invalid id"))
		return uuid.Nil, false
	}
	return id, true
}

func conversationResponse(c *models.Conversation) dtos.ConversationResponse {
	r := dtos.ConversationResponse{
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

func messageResponse(m *models.Message) dtos.MessageResponse {
	return dtos.MessageResponse{
		ID: m.ID.String(), Direction: m.Direction, SenderType: m.SenderType, Content: m.Content,
		Media: m.Media, ExternalMessageID: m.ExternalMessageID, Status: m.Status, CreatedAt: m.CreatedAt,
	}
}
