package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/dto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// UploadText handles POST /knowledge-bases/:id/documents/text.
// @Summary  Upload document (text)
// @Tags     knowledge
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    id   path string                  true "Knowledge base ID"
// @Param    body body dto.UploadTextRequest   true "Text content"
// @Success  201 {object} dto.DocumentResponse
// @Router   /knowledge-bases/{id}/documents/text [post]
func (h *Handler) UploadText(c *gin.Context) {
	kbID, ok := parseID(c, "id")
	if !ok {
		return
	}
	var in dto.UploadTextRequest
	if !httputil.BindJSON(c, &in) {
		return
	}
	doc, err := h.svc.UploadText(c.Request.Context(), kbID, in.Title, in.Content)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, documentResponse(doc))
}
