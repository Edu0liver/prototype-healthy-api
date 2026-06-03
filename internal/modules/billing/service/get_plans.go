package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
)

// ListPlans returns the active plan catalogue (system scope; plans are global).
func (s *Service) ListPlans(ctx context.Context) ([]models.Plan, error) {
	var plans []models.Plan
	if err := s.db.System(ctx, func(ctx context.Context) error {
		p, e := s.repo.ListActivePlans(ctx)
		plans = p
		return e
	}); err != nil {
		return nil, err
	}
	return plans, nil
}
