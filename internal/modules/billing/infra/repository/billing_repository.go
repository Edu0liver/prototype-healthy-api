// Package repository implements persistence for the billing module: plan +
// subscription reads, resource counts (for hard limits), usage-event writes and
// per-period usage aggregation.
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a subscription does not exist.
var ErrNotFound = errors.New("billing: not found")

// Repository persists billing data.
type Repository struct{}

// New builds the repository.
func New() *Repository { return &Repository{} }

// Limits is the flattened plan + subscription period for one company. Quota/cap
// of 0 means unlimited; overage cents of 0 means overage is disabled.
type Limits struct {
	PlanCode string `gorm:"column:code"`
	Status   string `gorm:"column:status"`

	QuotaAIMessages   int   `gorm:"column:quota_ai_messages"`
	QuotaTokens       int64 `gorm:"column:quota_tokens"`
	QuotaAudioMinutes int   `gorm:"column:quota_audio_minutes"`
	QuotaStorageMB    int   `gorm:"column:quota_storage_mb"`

	MaxChannels int `gorm:"column:max_channels"`
	MaxAgents   int `gorm:"column:max_agents"`
	MaxKB       int `gorm:"column:max_kb"`
	MaxSeats    int `gorm:"column:max_seats"`

	OveragePerMsgCents      int `gorm:"column:overage_per_msg_cents"`
	OveragePer1kTokensCents int `gorm:"column:overage_per_1k_tokens_cents"`

	PeriodStart time.Time `gorm:"column:current_period_start"`
	PeriodEnd   time.Time `gorm:"column:current_period_end"`
}

// Active reports whether the subscription entitles the tenant to use the app:
// status trialing/active and the current period has not lapsed.
func (l Limits) Active(now time.Time) bool {
	if l.Status != "trialing" && l.Status != "active" {
		return false
	}
	return l.PeriodEnd.IsZero() || l.PeriodEnd.After(now)
}

// LoadLimits joins subscriptions→plans for a company. plans/subscriptions are
// not under RLS, so the company_id filter is explicit. Runs in whatever
// tenant/system tx is already in context.
func (r *Repository) LoadLimits(ctx context.Context, companyID uuid.UUID) (*Limits, error) {
	var l Limits
	err := database.MustTx(ctx).
		Table("subscriptions AS s").
		Select("p.code, s.status, p.quota_ai_messages, p.quota_tokens, p.quota_audio_minutes, p.quota_storage_mb, "+
			"p.max_channels, p.max_agents, p.max_kb, p.max_seats, "+
			"p.overage_per_msg_cents, p.overage_per_1k_tokens_cents, "+
			"s.current_period_start, s.current_period_end").
		Joins("JOIN plans p ON p.id = s.plan_id").
		Where("s.company_id = ?", companyID).
		Take(&l).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &l, err
}

// CountResource counts rows of a tenant-scoped table for the caller's company.
// table must be a trusted constant (never user input).
func (r *Repository) CountResource(ctx context.Context, table string) (int64, error) {
	var n int64
	err := database.MustTx(ctx).Table(table).Scopes(database.TenantScope(ctx)).Count(&n).Error
	return n, err
}

// InsertUsageEvent appends one metering record (tenant-scoped; called inside a
// db.Tenant tx by the meter worker).
func (r *Repository) InsertUsageEvent(ctx context.Context, e *models.UsageEvent) error {
	return database.MustTx(ctx).Create(e).Error
}

// GetSubscription loads a company's subscription joined with its plan.
func (r *Repository) GetSubscription(ctx context.Context, companyID uuid.UUID) (*models.Subscription, *models.Plan, error) {
	var sub models.Subscription
	err := database.MustTx(ctx).Where("company_id = ?", companyID).Take(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, ErrNotFound
	}
	if err != nil {
		return nil, nil, err
	}
	var plan models.Plan
	if err := database.MustTx(ctx).Where("id = ?", sub.PlanID).Take(&plan).Error; err != nil {
		return nil, nil, err
	}
	return &sub, &plan, nil
}

// ListActivePlans returns the active plan catalogue ordered by price (system
// scope; plans are global and not under RLS).
func (r *Repository) ListActivePlans(ctx context.Context) ([]models.Plan, error) {
	var out []models.Plan
	err := database.MustTx(ctx).Where("is_active = ?", true).Order("price_cents ASC").Find(&out).Error
	return out, err
}

// GetPlanByCode loads a plan by its code (e.g. "pro").
func (r *Repository) GetPlanByCode(ctx context.Context, code string) (*models.Plan, error) {
	var p models.Plan
	err := database.MustTx(ctx).Where("code = ?", code).Take(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

// GetPlanByStripePrice maps a Stripe price id back to a local plan.
func (r *Repository) GetPlanByStripePrice(ctx context.Context, priceID string) (*models.Plan, error) {
	var p models.Plan
	err := database.MustTx(ctx).Where("stripe_price_id = ?", priceID).Take(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

// ActivateSubscription upserts the company's subscription on a completed
// checkout: it binds the plan + Stripe ids and marks it active. The period end
// is refined later by the customer.subscription.updated webhook. Runs in a
// system scope (the webhook has no tenant context).
func (r *Repository) ActivateSubscription(ctx context.Context, companyID, planID uuid.UUID, customerID, subscriptionID string) error {
	return database.MustTx(ctx).Exec(
		`INSERT INTO subscriptions
		   (id, company_id, plan_id, status, billing_cycle,
		    current_period_start, current_period_end,
		    stripe_customer_id, stripe_subscription_id, created_at, updated_at)
		 VALUES (gen_random_uuid(), ?, ?, 'active', 'monthly',
		    now(), now() + interval '1 month', ?, ?, now(), now())
		 ON CONFLICT (company_id) DO UPDATE SET
		    plan_id = EXCLUDED.plan_id,
		    status = 'active',
		    stripe_customer_id = EXCLUDED.stripe_customer_id,
		    stripe_subscription_id = EXCLUDED.stripe_subscription_id,
		    updated_at = now()`,
		companyID, planID, nullable(customerID), nullable(subscriptionID),
	).Error
}

// UpdateSubscriptionByStripeID applies a subscription lifecycle change keyed by
// the Stripe subscription id. planID is optional (set on plan changes); a zero
// UUID leaves the plan untouched. periodEndUnix of 0 leaves the period end as-is.
func (r *Repository) UpdateSubscriptionByStripeID(ctx context.Context, subscriptionID, status string, periodEndUnix int64, cancelAtEnd bool, planID uuid.UUID) (int64, error) {
	updates := map[string]any{
		"status":               status,
		"cancel_at_period_end": cancelAtEnd,
		"updated_at":           gorm.Expr("now()"),
	}
	if periodEndUnix > 0 {
		updates["current_period_end"] = time.Unix(periodEndUnix, 0).UTC()
	}
	if planID != uuid.Nil {
		updates["plan_id"] = planID
	}
	res := database.MustTx(ctx).Table("subscriptions").
		Where("stripe_subscription_id = ?", subscriptionID).
		Updates(updates)
	return res.RowsAffected, res.Error
}

// CompanyIDByStripeSubscription resolves the tenant owning a Stripe subscription
// (used to invalidate that tenant's cached limits after a webhook).
func (r *Repository) CompanyIDByStripeSubscription(ctx context.Context, subscriptionID string) (uuid.UUID, error) {
	var sub models.Subscription
	err := database.MustTx(ctx).Select("company_id").
		Where("stripe_subscription_id = ?", subscriptionID).Take(&sub).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return uuid.Nil, ErrNotFound
	}
	return sub.CompanyID, err
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// KindSum is one aggregated usage line.
type KindSum struct {
	Kind  string `gorm:"column:kind"`
	Total int64  `gorm:"column:total"`
}

// SumUsageSince aggregates usage quantity by kind since a timestamp
// (tenant-scoped read for the panel's usage dashboard).
func (r *Repository) SumUsageSince(ctx context.Context, since time.Time) ([]KindSum, error) {
	var out []KindSum
	err := database.MustTx(ctx).
		Table("usage_events").
		Scopes(database.TenantScope(ctx)).
		Select("kind, COALESCE(SUM(quantity),0) AS total").
		Where("created_at >= ?", since).
		Group("kind").
		Scan(&out).Error
	return out, err
}
