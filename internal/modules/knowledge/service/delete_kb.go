package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	"github.com/google/uuid"
)

// DeleteKB removes a knowledge base.
func (s *Service) DeleteKB(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteKB(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrKBNotFound
		}
		return err
	}
	return nil
}
