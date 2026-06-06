package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/google/uuid"
)

// Repository is the persistence contract consumed by the iam service.
type Repository interface {
	FindByEmailGlobal(ctx context.Context, email string) (userID, companyID uuid.UUID, err error)
	FindRoleByName(ctx context.Context, name string) (*models.SystemRole, error)
	Create(ctx context.Context, u *models.User) error
	Update(ctx context.Context, u *models.User) error
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	CountUsers(ctx context.Context) (int64, error)
	ListByCompany(ctx context.Context) ([]models.User, error)
}

// Mailer sends transactional email (satisfied by platform/mailer).
type Mailer interface {
	Send(to, subject, html string) error
}

// QuotaGuard enforces plan resource caps at create time (billing module).
type QuotaGuard interface {
	EnsureResource(ctx context.Context, companyID uuid.UUID, resource string) error
}

// noopQuota is the default guard (no enforcement) used until WithBilling runs.
type noopQuota struct{}

func (noopQuota) EnsureResource(context.Context, uuid.UUID, string) error { return nil }
