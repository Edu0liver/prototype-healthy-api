package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
)

// List returns all channels in the tenant.
func (s *Service) List(ctx context.Context) ([]models.Channel, error) {
	return s.repo.List(ctx)
}
