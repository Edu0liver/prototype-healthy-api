package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	"github.com/google/uuid"
)

// GetKB loads a knowledge base by id.
func (s *Service) GetKB(ctx context.Context, id uuid.UUID) (*models.KnowledgeBase, error) {
	kb, err := s.repo.GetKB(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrKBNotFound
	}
	return kb, err
}
