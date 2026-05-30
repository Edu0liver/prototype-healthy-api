package services

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/google/uuid"
)

// Repository is the persistence contract consumed by the tenant service.
type Repository interface {
	CreateCompany(ctx context.Context, c *models.Company) error
	GetCompany(ctx context.Context, id uuid.UUID) (*models.Company, error)
	FindCompanyByDomain(ctx context.Context, domain string) (*models.Company, error)
	SlugExists(ctx context.Context, slug string) (bool, error)
	UpsertBranding(ctx context.Context, b *models.CompanyBranding) error
	GetBranding(ctx context.Context, companyID uuid.UUID) (*models.CompanyBranding, error)
	AddDomain(ctx context.Context, d *models.CompanyDomain) error
	ListDomains(ctx context.Context, companyID uuid.UUID) ([]models.CompanyDomain, error)
}
