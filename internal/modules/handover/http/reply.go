package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Reply handles POST /conversations/:id/handover/reply.
func (h *Handler) Reply(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var in dto.ReplyRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	if err := h.svc.Reply(c.Request.Context(), id, in.Content); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}
