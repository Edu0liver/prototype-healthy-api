package http

import (
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// List handles GET /conversations with optional filters (RF-LOG-03).
// @Summary  List conversations
// @Tags     conversations
// @Security BearerAuth
// @Produce  json
// @Param    state      query string false "Filter by state (open/closed/human)"
// @Param    channel_id query string false "Filter by channel UUID"
// @Param    since      query string false "Filter by opened_at >= RFC3339 timestamp"
// @Success  200 {object} map[string][]dto.ConversationResponse
// @Router   /conversations [get]
func (h *Handler) List(c *gin.Context) {
	f := repository.ConversationFilter{State: c.Query("state")}
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
	out := make([]dto.ConversationResponse, len(items))
	for i := range items {
		out[i] = conversationResponse(&items[i])
	}
	httputil.OK(c, gin.H{"conversations": out})
}

// Get handles GET /conversations/:id.
// @Summary  Get conversation
// @Tags     conversations
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Conversation ID"
// @Success  200 {object} dto.ConversationResponse
// @Router   /conversations/{id} [get]
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
// @Summary  List messages
// @Tags     conversations
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Conversation ID"
// @Success  200 {object} map[string][]dto.MessageResponse
// @Router   /conversations/{id}/messages [get]
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
	out := make([]dto.MessageResponse, len(msgs))
	for i := range msgs {
		out[i] = messageResponse(&msgs[i])
	}
	httputil.OK(c, gin.H{"messages": out})
}
