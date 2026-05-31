package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/google/uuid"
)

// UploadFile stores an uploaded file and kicks off async indexing (RF-RAG-02).
func (s *Service) UploadFile(ctx context.Context, kbID uuid.UUID, filename string, data []byte) (*models.Document, error) {
	return s.createAndIngest(ctx, kbID, "file", filename, data)
}
