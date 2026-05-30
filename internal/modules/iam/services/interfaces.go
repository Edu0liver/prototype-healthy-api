package services

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/google/uuid"
)

// Repository is the persistence contract consumed by the iam service.
type Repository interface {
	CompanyIDBySlug(ctx context.Context, slug string) (uuid.UUID, error)
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
