package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// CreateKB handles POST /knowledge-bases.
// @Summary  Create knowledge base
// @Tags     knowledge
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.CreateKBRequest true "Knowledge base"
// @Success  201 {object} dto.KBResponse
// @Failure  400 {object} httputil.ErrorResponse "invalid request"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /knowledge-bases [post]
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
