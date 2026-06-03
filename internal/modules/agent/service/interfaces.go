package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/google/uuid"
)

// Repository is the persistence contract consumed by the agent service.
type Repository interface {
	Create(ctx context.Context, a *models.Agent) error
	Update(ctx context.Context, a *models.Agent) error
	Get(ctx context.Context, id uuid.UUID) (*models.Agent, error)
	List(ctx context.Context) ([]models.Agent, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// QuotaGuard enforces a plan's resource caps at create time (billing module).
type QuotaGuard interface {
	EnsureResource(ctx context.Context, companyID uuid.UUID, resource string) error
}

// noopQuota is the default guard (no enforcement) used until WithBilling runs.
type noopQuota struct{}

func (noopQuota) EnsureResource(context.Context, uuid.UUID, string) error { return nil }
