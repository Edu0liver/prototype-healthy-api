package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/repository"
	"github.com/google/uuid"
)

func TestGetSubscription_PassThrough(t *testing.T) {
	company := uuid.New()
	wantSub := &models.Subscription{CompanyID: company, Status: models.StatusActive}
	wantPlan := &models.Plan{Code: "pro"}
	svc := &Service{repo: &mockRepo{
		getSubFn: func(_ context.Context, id uuid.UUID) (*models.Subscription, *models.Plan, error) {
			if id != company {
				t.Fatalf("company mismatch: %s != %s", id, company)
			}
			return wantSub, wantPlan, nil
		},
	}}

	sub, plan, err := svc.GetSubscription(context.Background(), company)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub != wantSub || plan != wantPlan {
		t.Fatal("expected the repository's subscription/plan to pass through")
	}
}

func TestGetSubscription_NotFoundMapped(t *testing.T) {
	svc := &Service{repo: &mockRepo{
		getSubFn: func(context.Context, uuid.UUID) (*models.Subscription, *models.Plan, error) {
			return nil, nil, repository.ErrNotFound
		},
	}}

	_, _, err := svc.GetSubscription(context.Background(), uuid.New())
	if !errors.Is(err, ErrSubscriptionNotFn) {
		t.Fatalf("expected ErrSubscriptionNotFn, got %v", err)
	}
}
