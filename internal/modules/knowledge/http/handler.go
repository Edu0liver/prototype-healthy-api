// Package http exposes the knowledge module's Gin handlers (split per resource).
package http

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/service"
	"github.com/Edu0liver/prototype-healthy-api/pkg/httputil"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// maxUploadBytes caps a single document upload.
const maxUploadBytes = 20 << 20 // 20 MiB

// Handler serves knowledge-base, document and agent-link endpoints.
type Handler struct {
	svc *service.Service
}

// NewHandler builds the handler.
func NewHandler(svc *service.Service) *Handler { return &Handler{svc: svc} }

func parseID(c *gin.Context, name string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(name))
	if err != nil {
		httputil.Fail(c, httputil.BadRequest("invalid "+name))
		return uuid.Nil, false
	}
	return id, true
}

func kbResponse(kb *models.KnowledgeBase) dto.KBResponse {
	return dto.KBResponse{
		ID: kb.ID.String(), Name: kb.Name, Description: kb.Description,
		EmbeddingModel: kb.EmbeddingModel, ChunkSize: kb.ChunkSize, ChunkOverlap: kb.ChunkOverlap,
	}
}

func documentResponse(d *models.Document) dto.DocumentResponse {
	return dto.DocumentResponse{
		ID: d.ID.String(), Filename: d.Filename, SourceType: d.SourceType,
		Status: d.Status, Error: d.Error, TokenCount: d.TokenCount, CreatedAt: d.CreatedAt,
	}
}
