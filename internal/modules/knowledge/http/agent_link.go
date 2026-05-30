package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// LinkAgent handles POST /agents/:id/knowledge-bases/:kbId.
func (h *Handler) LinkAgent(c *gin.Context) {
	agentID, ok := parseID(c, "id")
	if !ok {
		return
	}
	kbID, ok := parseID(c, "kbId")
	if !ok {
		return
	}
	if err := h.svc.LinkAgent(c.Request.Context(), agentID, kbID); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}

// UnlinkAgent handles DELETE /agents/:id/knowledge-bases/:kbId.
func (h *Handler) UnlinkAgent(c *gin.Context) {
	agentID, ok := parseID(c, "id")
	if !ok {
		return
	}
	kbID, ok := parseID(c, "kbId")
	if !ok {
		return
	}
	if err := h.svc.UnlinkAgent(c.Request.Context(), agentID, kbID); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}
