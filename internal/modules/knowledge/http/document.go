package http

import (
	"io"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/dto"
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
// @Success  201 {object} dto.DocumentResponse
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

// ListDocuments handles GET /knowledge-bases/:id/documents.
// @Summary  List documents
// @Tags     knowledge
// @Security BearerAuth
// @Produce  json
// @Param    id path string true "Knowledge base ID"
// @Success  200 {object} map[string][]dto.DocumentResponse
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

// DeleteDocument handles DELETE /documents/:docId.
// @Summary  Delete document
// @Tags     knowledge
// @Security BearerAuth
// @Param    docId path string true "Document ID"
// @Success  204
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
