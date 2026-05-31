package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/google/uuid"
)

// ListDocuments returns documents of a knowledge base.
func (s *Service) ListDocuments(ctx context.Context, kbID uuid.UUID) ([]models.Document, error) {
	return s.repo.ListDocuments(ctx, kbID)
}
