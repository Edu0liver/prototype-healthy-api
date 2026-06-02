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
// @Failure  400 {object} httputil.ErrorResponse "invalid query filter"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
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
