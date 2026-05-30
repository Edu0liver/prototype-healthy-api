package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/repository"
	"github.com/google/uuid"
)

// Delete removes an agent.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrAgentNotFound
		}
		return err
	}
	return nil
}
