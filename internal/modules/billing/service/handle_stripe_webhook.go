package service

import (
	"context"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/pkg/stripe"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// HandleWebhook verifies and processes a Stripe webhook. It is idempotent
// (dedupe by event id) and runs in a system scope — the request carries no
// tenant context. Unhandled event types are acknowledged and ignored.
func (s *Service) HandleWebhook(ctx context.Context, payload []byte, sigHeader string) error {
	if s.stripe == nil {
		return ErrStripeDisabled
	}
	evt, err := s.stripe.VerifyWebhook(payload, sigHeader)
	if err != nil {
		return ErrInvalidSignature
	}

	// Idempotency: process each Stripe event id at most once.
	if s.rdb != nil {
		first, ferr := s.rdb.FirstSeen(ctx, "stripe:evt:"+evt.ID, 24*time.Hour)
		if ferr == nil && !first {
			s.log.Debug("billing: duplicate stripe event, skipping", zap.String("event_id", evt.ID))
			return nil
		}
	}

	obj := evt.Data.Object
	s.log.Debug("billing: stripe event", zap.String("type", evt.Type), zap.String("event_id", evt.ID))

	switch evt.Type {
	case "checkout.session.completed":
		return s.onCheckoutCompleted(ctx, obj)
	case "customer.subscription.updated":
		return s.onSubscriptionChanged(ctx, obj.ID, mapStatus(obj.Status), obj.CurrentPeriodEnd, obj.CancelAtPeriodEnd, obj.FirstPriceID())
	case "customer.subscription.deleted":
		return s.onSubscriptionChanged(ctx, obj.ID, models.StatusCanceled, obj.CurrentPeriodEnd, obj.CancelAtPeriodEnd, "")
	case "invoice.payment_failed":
		return s.onSubscriptionChanged(ctx, obj.Subscription, models.StatusPastDue, 0, false, "")
	default:
		s.log.Debug("billing: unhandled stripe event type", zap.String("type", evt.Type))
		return nil
	}
}

func (s *Service) onCheckoutCompleted(ctx context.Context, obj stripe.EventObject) error {
	companyID, err := uuid.Parse(firstNonEmpty(obj.ClientReferenceID, obj.Metadata["company_id"]))
	if err != nil {
		s.log.Warn("billing: checkout.completed without valid company id", zap.String("ref", obj.ClientReferenceID))
		return nil
	}
	planID, err := uuid.Parse(obj.Metadata["plan_id"])
	if err != nil {
		s.log.Warn("billing: checkout.completed without valid plan id")
		return nil
	}
	if err := s.db.System(ctx, func(ctx context.Context) error {
		return s.repo.ActivateSubscription(ctx, companyID, planID, obj.Customer, obj.Subscription)
	}); err != nil {
		return err
	}
	s.InvalidateLimits(ctx, companyID)
	s.log.Info("billing: subscription activated", zap.String("company_id", companyID.String()), zap.String("plan_id", planID.String()))
	return nil
}

func (s *Service) onSubscriptionChanged(ctx context.Context, subscriptionID, status string, periodEndUnix int64, cancelAtEnd bool, priceID string) error {
	if subscriptionID == "" {
		return nil
	}
	return s.db.System(ctx, func(ctx context.Context) error {
		// Resolve a plan change from the price, if the event carried one.
		planID := uuid.Nil
		if priceID != "" {
			if p, err := s.repo.GetPlanByStripePrice(ctx, priceID); err == nil {
				planID = p.ID
			}
		}
		rows, err := s.repo.UpdateSubscriptionByStripeID(ctx, subscriptionID, status, periodEndUnix, cancelAtEnd, planID)
		if err != nil {
			return err
		}
		if rows == 0 {
			s.log.Warn("billing: subscription webhook for unknown stripe id", zap.String("sub_id", subscriptionID))
			return nil
		}
		if companyID, err := s.repo.CompanyIDByStripeSubscription(ctx, subscriptionID); err == nil {
			s.InvalidateLimits(ctx, companyID)
		}
		return nil
	})
}

// mapStatus normalizes Stripe subscription statuses onto the local set.
func mapStatus(stripeStatus string) string {
	switch stripeStatus {
	case "active", "trialing", "past_due", "canceled":
		return stripeStatus
	case "unpaid":
		return models.StatusPastDue
	case "incomplete", "incomplete_expired", "paused":
		return models.StatusSuspended
	default:
		return models.StatusActive
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
