package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/google/uuid"
)

// Get returns an automation by id.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*models.Automation, error) {
	return s.get(ctx, id)
}
