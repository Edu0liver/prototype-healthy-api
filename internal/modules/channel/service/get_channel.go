package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/google/uuid"
)

// Get returns a channel by id.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	return s.get(ctx, id)
}

// List returns all channels in the tenant.
func (s *Service) List(ctx context.Context) ([]models.Channel, error) {
	return s.repo.List(ctx)
}
