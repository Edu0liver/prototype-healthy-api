package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
)

// List returns all automations in the tenant.
func (s *Service) List(ctx context.Context) ([]models.Automation, error) {
	return s.repo.List(ctx)
}
