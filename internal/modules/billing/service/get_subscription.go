package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/repository"
	"github.com/google/uuid"
)

// GetSubscription returns the caller tenant's subscription joined with its plan.
func (s *Service) GetSubscription(ctx context.Context, companyID uuid.UUID) (*models.Subscription, *models.Plan, error) {
	sub, plan, err := s.repo.GetSubscription(ctx, companyID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, nil, ErrSubscriptionNotFn
	}
	return sub, plan, err
}
