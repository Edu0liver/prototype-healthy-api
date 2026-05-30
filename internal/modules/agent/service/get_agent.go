package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/google/uuid"
)

// Get returns an agent by id.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*models.Agent, error) {
	return s.get(ctx, id)
}
