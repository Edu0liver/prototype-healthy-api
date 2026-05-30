package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// Take handles POST /conversations/:id/handover/take.
// @Summary  Take conversation (human)
// @Tags     handover
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Conversation ID"
// @Success  200 {object} map[string]string
// @Router   /conversations/{id}/handover/take [post]
func (h *Handler) Take(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.Take(c.Request.Context(), id, appctx.UserID(c.Request.Context())); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, gin.H{"state": "human"})
}
