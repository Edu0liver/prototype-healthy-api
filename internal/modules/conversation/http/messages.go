package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

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
