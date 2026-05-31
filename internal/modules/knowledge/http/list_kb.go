package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// ListKB handles GET /knowledge-bases.
// @Summary  List knowledge bases
// @Tags     knowledge
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} map[string][]dto.KBResponse
// @Router   /knowledge-bases [get]
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
