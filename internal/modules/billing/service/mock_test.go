package service

import (
	"context"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/repository"
	"github.com/google/uuid"
)

// mockRepo is a hand-written stub of the service.Repository contract; each test
// sets only the function it needs.
type mockRepo struct {
	loadLimitsFn     func(ctx context.Context, companyID uuid.UUID) (*repository.Limits, error)
	countFn          func(ctx context.Context, table string) (int64, error)
	insertFn         func(ctx context.Context, e *models.UsageEvent) error
	getSubFn         func(ctx context.Context, companyID uuid.UUID) (*models.Subscription, *models.Plan, error)
	sumUsageFn       func(ctx context.Context, since time.Time) ([]repository.KindSum, error)
	getPlanByCodeFn  func(ctx context.Context, code string) (*models.Plan, error)
	getPlanByPriceFn func(ctx context.Context, priceID string) (*models.Plan, error)
	activateFn       func(ctx context.Context, companyID, planID uuid.UUID, customerID, subscriptionID string) error
	updateBySubFn    func(ctx context.Context, subscriptionID, status string, periodEndUnix int64, cancelAtEnd bool, planID uuid.UUID) (int64, error)
	companyBySubFn   func(ctx context.Context, subscriptionID string) (uuid.UUID, error)
}

func (m *mockRepo) LoadLimits(ctx context.Context, companyID uuid.UUID) (*repository.Limits, error) {
	return m.loadLimitsFn(ctx, companyID)
}
func (m *mockRepo) CountResource(ctx context.Context, table string) (int64, error) {
	return m.countFn(ctx, table)
}
func (m *mockRepo) InsertUsageEvent(ctx context.Context, e *models.UsageEvent) error {
	return m.insertFn(ctx, e)
}
func (m *mockRepo) GetSubscription(ctx context.Context, companyID uuid.UUID) (*models.Subscription, *models.Plan, error) {
	return m.getSubFn(ctx, companyID)
}
func (m *mockRepo) SumUsageSince(ctx context.Context, since time.Time) ([]repository.KindSum, error) {
	return m.sumUsageFn(ctx, since)
}
func (m *mockRepo) GetPlanByCode(ctx context.Context, code string) (*models.Plan, error) {
	return m.getPlanByCodeFn(ctx, code)
}
func (m *mockRepo) GetPlanByStripePrice(ctx context.Context, priceID string) (*models.Plan, error) {
	return m.getPlanByPriceFn(ctx, priceID)
}
func (m *mockRepo) ActivateSubscription(ctx context.Context, companyID, planID uuid.UUID, customerID, subscriptionID string) error {
	return m.activateFn(ctx, companyID, planID, customerID, subscriptionID)
}
func (m *mockRepo) UpdateSubscriptionByStripeID(ctx context.Context, subscriptionID, status string, periodEndUnix int64, cancelAtEnd bool, planID uuid.UUID) (int64, error) {
	return m.updateBySubFn(ctx, subscriptionID, status, periodEndUnix, cancelAtEnd, planID)
}
func (m *mockRepo) CompanyIDByStripeSubscription(ctx context.Context, subscriptionID string) (uuid.UUID, error) {
	return m.companyBySubFn(ctx, subscriptionID)
}
