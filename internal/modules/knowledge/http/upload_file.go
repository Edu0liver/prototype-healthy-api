package http

import (
	"io"

	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
)

// UploadFile handles POST /knowledge-bases/:id/documents (multipart).
// @Summary  Upload document (file)
// @Tags     knowledge
// @Security BearerAuth
// @Accept   multipart/form-data
// @Produce  json
// @Param    id   path     string true "Knowledge base ID"
// @Param    file formData file   true "File (PDF, DOCX, TXT, etc.)"
// @Success  201
// @Failure  400 {object} httputil.ErrorResponse "invalid id or missing file"
// @Failure  401 {object} httputil.ErrorResponse "missing or invalid token"
// @Failure  403 {object} httputil.ErrorResponse "insufficient role"
// @Failure  404 {object} httputil.ErrorResponse "knowledge base not found"
// @Failure  429 {object} httputil.ErrorResponse "rate limit exceeded"
// @Failure  500 {object} httputil.ErrorResponse "internal error"
// @Router   /knowledge-bases/{id}/documents [post]
func (h *Handler) UploadFile(c *gin.Context) {
	kbID, ok := parseID(c, "id")
	if !ok {
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		httputil.Fail(c, httputil.BadRequest("missing file field"))
		return
	}
	if fileHeader.Size > maxUploadBytes {
		httputil.Fail(c, httputil.BadRequest("file too large"))
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxUploadBytes))
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	doc, err := h.svc.UploadFile(c.Request.Context(), kbID, fileHeader.Filename, data)
	if err != nil {
		httputil.Fail(c, err)
		return
	}
	httputil.Created(c, documentResponse(doc))
}
