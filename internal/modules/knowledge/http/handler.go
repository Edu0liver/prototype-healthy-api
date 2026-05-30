// Package http exposes the knowledge module's Gin handlers.
package http

import (
	"io"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/dtos"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/services"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// maxUploadBytes caps a single document upload.
const maxUploadBytes = 20 << 20 // 20 MiB

// Handler serves knowledge-base, document and agent-link endpoints.
type Handler struct {
	svc *services.Service
}

// NewHandler builds the handler.
func NewHandler(svc *services.Service) *Handler { return &Handler{svc: svc} }

// CreateKB handles POST /knowledge-bases.
func (h *Handler) CreateKB(c *gin.Context) {
	var in dtos.CreateKBRequest
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
	out := make([]dtos.KBResponse, len(kbs))
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

// UploadFile handles POST /knowledge-bases/:id/documents (multipart).
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
func (h *Handler) UploadText(c *gin.Context) {
	kbID, ok := parseID(c, "id")
	if !ok {
		return
	}
	var in dtos.UploadTextRequest
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
	out := make([]dtos.DocumentResponse, len(docs))
	for i := range docs {
		out[i] = documentResponse(&docs[i])
	}
	httputil.OK(c, gin.H{"documents": out})
}

// DeleteDocument handles DELETE /documents/:docId.
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

// LinkAgent handles POST /agents/:agentId/knowledge-bases/:kbId.
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

// UnlinkAgent handles DELETE /agents/:agentId/knowledge-bases/:kbId.
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

func parseID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		httputil.Fail(c, httputil.BadRequest("invalid "+name))
		return uuid.Nil, false
	}
	return id, true
}

func kbResponse(kb *models.KnowledgeBase) dtos.KBResponse {
	return dtos.KBResponse{
		ID: kb.ID.String(), Name: kb.Name, Description: kb.Description,
		EmbeddingModel: kb.EmbeddingModel, ChunkSize: kb.ChunkSize, ChunkOverlap: kb.ChunkOverlap,
	}
}

func documentResponse(d *models.Document) dtos.DocumentResponse {
	return dtos.DocumentResponse{
		ID: d.ID.String(), Filename: d.Filename, SourceType: d.SourceType,
		Status: d.Status, Error: d.Error, TokenCount: d.TokenCount, CreatedAt: d.CreatedAt,
	}
}
