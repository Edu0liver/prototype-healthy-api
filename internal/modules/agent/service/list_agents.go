package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
)

// List returns all agents in the tenant.
func (s *Service) List(ctx context.Context) ([]models.Agent, error) {
	return s.repo.List(ctx)
}
