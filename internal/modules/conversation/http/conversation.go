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
	out := make([]dto.MessageResponse, len(msgs))
	for i := range msgs {
		out[i] = messageResponse(&msgs[i])
	}
	httputil.OK(c, gin.H{"messages": out})
}
