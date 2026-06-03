package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/pkg/stripe"
	"github.com/google/uuid"
)

// CreateCheckout starts a Stripe Checkout Session for the given plan and returns
// the hosted payment URL. The subscription is only persisted once Stripe calls
// back on checkout.session.completed (see HandleWebhook).
func (s *Service) CreateCheckout(ctx context.Context, companyID uuid.UUID, planCode, customerEmail string) (string, error) {
	if s.stripe == nil {
		return "", ErrStripeDisabled
	}

	var plan *models.Plan
	if err := s.db.System(ctx, func(ctx context.Context) error {
		p, e := s.repo.GetPlanByCode(ctx, planCode)
		plan = p
		return e
	}); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrPlanNotFound
		}
		return "", err
	}
	if plan.StripePriceID == "" {
		return "", ErrPlanNotPurchasable
	}

	sess, err := s.stripe.CreateCheckoutSession(ctx, stripe.CheckoutParams{
		PriceID:       plan.StripePriceID,
		CompanyID:     companyID.String(),
		CustomerEmail: customerEmail,
		Metadata: map[string]string{
			"plan_id":   plan.ID.String(),
			"plan_code": plan.Code,
		},
	})
	if err != nil {
		return "", err
	}
	return sess.URL, nil
}
