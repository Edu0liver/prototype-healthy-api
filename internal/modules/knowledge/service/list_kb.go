package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
)

// ListKB returns all knowledge bases in the tenant.
func (s *Service) ListKB(ctx context.Context) ([]models.KnowledgeBase, error) {
	return s.repo.ListKB(ctx)
}
