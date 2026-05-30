package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// CreateKB handles POST /knowledge-bases.
func (h *Handler) CreateKB(c *gin.Context) {
	var in dto.CreateKBRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	kb, err := h.svc.CreateKB(c.Request.Context(), in)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, kbResponse(kb))
}

// ListKB handles GET /knowledge-bases.
func (h *Handler) ListKB(c *gin.Context) {
	kbs, err := h.svc.ListKB(c.Request.Context())
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dto.KBResponse, len(kbs))
	for i := range kbs {
		out[i] = kbResponse(&kbs[i])
	}
	httputil.OK(c, gin.H{"knowledge_bases": out})
}

// GetKB handles GET /knowledge-bases/:id.
func (h *Handler) GetKB(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	kb, err := h.svc.GetKB(c.Request.Context(), id)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.OK(c, kbResponse(kb))
}

// DeleteKB handles DELETE /knowledge-bases/:id.
func (h *Handler) DeleteKB(c *gin.Context) {
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteKB(c.Request.Context(), id); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}
