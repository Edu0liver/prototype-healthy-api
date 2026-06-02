package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// ListDocuments handles GET /knowledge-bases/:id/documents.
// @Summary  List documents
// @Tags     knowledge
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Knowledge base ID"
// @Success  200 {object} map[string][]dto.DocumentResponse
// @Failure  400 {object} httputil.ErrorResponse "invalid id"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  404 {object} httputil.ErrorResponse "knowledge base not found"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /knowledge-bases/{id}/documents [get]
func (h *Handler) ListDocuments(c *gin.Context) {
	kbID, ok := parseID(c, "id")
	if !ok {
		return
	}
	docs, err := h.svc.ListDocuments(c.Request.Context(), kbID)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	out := make([]dto.DocumentResponse, len(docs))
	for i := range docs {
		out[i] = documentResponse(&docs[i])
	}
	httputil.OK(c, gin.H{"documents": out})
}
