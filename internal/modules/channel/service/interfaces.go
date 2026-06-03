package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/google/uuid"
)

// Repository is the persistence contract consumed by the channel service.
type Repository interface {
	Create(ctx context.Context, c *models.Channel) error
	Update(ctx context.Context, c *models.Channel) error
	Get(ctx context.Context, id uuid.UUID) (*models.Channel, error)
	List(ctx context.Context) ([]models.Channel, error)
	// ListAllActive returns connected/connecting channels across all tenants.
	// Must be called within a db.System() scope.
	ListAllActive(ctx context.Context) ([]models.Channel, error)
}

// QuotaGuard enforces a plan's resource caps at create time (billing module).
type QuotaGuard interface {
	EnsureResource(ctx context.Context, companyID uuid.UUID, resource string) error
}

// noopQuota is the default guard (no enforcement) used until WithBilling runs.
type noopQuota struct{}

func (noopQuota) EnsureResource(context.Context, uuid.UUID, string) error { return nil }
