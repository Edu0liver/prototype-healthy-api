package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/repository"
	"github.com/google/uuid"
)

// CreatePortalSession creates a Stripe Billing Portal session for the tenant
// and returns the hosted portal URL. The tenant must have an active Stripe
// customer id (i.e. have completed at least one checkout).
func (s *Service) CreatePortalSession(ctx context.Context, companyID uuid.UUID) (string, error) {
	if s.stripe == nil {
		return "", ErrStripeDisabled
	}

	var sub *models.Subscription
	if err := s.db.System(ctx, func(ctx context.Context) error {
		var e error
		sub, _, e = s.repo.GetSubscription(ctx, companyID)
		return e
	}); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrSubscriptionNotFn
		}
		return "", err
	}
	if sub.StripeCustomerID == "" {
		return "", ErrStripeDisabled
	}

	sess, err := s.stripe.CreatePortalSession(ctx, sub.StripeCustomerID)
	if err != nil {
		return "", err
	}
	return sess.URL, nil
}
