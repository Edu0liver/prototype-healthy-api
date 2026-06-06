package service

import (
	"context"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/pkg/stripe"
	"github.com/google/uuid"
)

// Repository is the persistence contract consumed by the billing service.
type Repository interface {
	LoadLimits(ctx context.Context, companyID uuid.UUID) (*repository.Limits, error)
	CountResource(ctx context.Context, table string) (int64, error)
	InsertUsageEvent(ctx context.Context, e *models.UsageEvent) error
	GetSubscription(ctx context.Context, companyID uuid.UUID) (*models.Subscription, *models.Plan, error)
	SumUsageSince(ctx context.Context, since time.Time) ([]repository.KindSum, error)

	ListActivePlans(ctx context.Context) ([]models.Plan, error)

	// Stripe gateway flow.
	GetPlanByCode(ctx context.Context, code string) (*models.Plan, error)
	GetPlanByStripePrice(ctx context.Context, priceID string) (*models.Plan, error)
	ActivateSubscription(ctx context.Context, companyID, planID uuid.UUID, customerID, subscriptionID string) error
	UpdateSubscriptionByStripeID(ctx context.Context, subscriptionID, status string, periodEndUnix int64, cancelAtEnd bool, planID uuid.UUID) (int64, error)
	CompanyIDByStripeSubscription(ctx context.Context, subscriptionID string) (uuid.UUID, error)
}

// StripeGateway is the billing gateway contract (satisfied by pkg/stripe.Client).
type StripeGateway interface {
	CreateCheckoutSession(ctx context.Context, p stripe.CheckoutParams) (*stripe.CheckoutSession, error)
	CreatePortalSession(ctx context.Context, customerID string) (*stripe.PortalSession, error)
	VerifyWebhook(payload []byte, sigHeader string) (*stripe.Event, error)
}
