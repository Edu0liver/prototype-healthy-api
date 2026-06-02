package http

import (
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// DeleteDocument handles DELETE /documents/:docId.
// @Summary  Delete document
// @Tags     knowledge
// @Security BearerAuth
// @Param    docId path string true "Document ID"
// @Success  204
// @Failure  400 {object} httputil.ErrorResponse "invalid id"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  404 {object} httputil.ErrorResponse "document not found"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /documents/{docId} [delete]
func (h *Handler) DeleteDocument(c *gin.Context) {
	id, ok := parseID(c, "docId")
	if !ok {
		return
	}
	if err := h.svc.DeleteDocument(c.Request.Context(), id); err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.NoContent(c)
}
