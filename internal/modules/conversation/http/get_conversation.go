package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Get handles GET /conversations/:id.
// @Summary  Get conversation
// @Tags     conversations
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Conversation ID"
// @Success  200
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
