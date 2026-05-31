// Package repositories implements persistence for the tenant module.
// The tenant registry tables (companies/company_domains/company_branding) are
// NOT under RLS, so isolation here is enforced by explicit lookup keys.
package repository

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrNotFound is returned when a row does not exist.
var ErrNotFound = errors.New("tenant: not found")

// Repository persists tenant entities.
type Repository struct{}

// New builds the repository.
func New() *Repository { return &Repository{} }

func wrap(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

// CreateCompany inserts a company plus its branding row.
func (r *Repository) CreateCompany(ctx context.Context, c *models.Company) error {
	return wrap(database.MustTx(ctx).Create(c).Error)
}

// GetCompany loads a company by id.
func (r *Repository) GetCompany(ctx context.Context, id uuid.UUID) (*models.Company, error) {
	var c models.Company
	if err := database.MustTx(ctx).First(&c, "id = ?", id).Error; err != nil {
		return nil, wrap(err)
	}
	return &c, nil
}

// FindCompanyByDomain resolves a Host to its company.
func (r *Repository) FindCompanyByDomain(ctx context.Context, domain string) (*models.Company, error) {
	var d models.CompanyDomain
	if err := database.MustTx(ctx).First(&d, "domain = ?", domain).Error; err != nil {
		return nil, wrap(err)
	}
	return r.GetCompany(ctx, d.CompanyID)
}

// SlugExists reports whether a slug is taken.
func (r *Repository) SlugExists(ctx context.Context, slug string) (bool, error) {
	var n int64
	err := database.MustTx(ctx).Model(&models.Company{}).Where("slug = ?", slug).Count(&n).Error
	return n > 0, err
}

// UpsertBranding creates or updates the branding row for a company.
func (r *Repository) UpsertBranding(ctx context.Context, b *models.CompanyBranding) error {
	return wrap(database.MustTx(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "company_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"logo_url", "favicon_url", "primary_color",
				"secondary_color", "email_sender_name", "updated_at",
			}),
		}).
		Create(b).Error)
}

// GetBranding loads the branding for a company.
func (r *Repository) GetBranding(ctx context.Context, companyID uuid.UUID) (*models.CompanyBranding, error) {
	var b models.CompanyBranding
	if err := database.MustTx(ctx).First(&b, "company_id = ?", companyID).Error; err != nil {
		return nil, wrap(err)
	}
	return &b, nil
}

// AddDomain inserts a domain mapping.
func (r *Repository) AddDomain(ctx context.Context, d *models.CompanyDomain) error {
	return wrap(database.MustTx(ctx).Create(d).Error)
}

// ListDomains returns all domains for a company.
func (r *Repository) ListDomains(ctx context.Context, companyID uuid.UUID) ([]models.CompanyDomain, error) {
	var out []models.CompanyDomain
	err := database.MustTx(ctx).Where("company_id = ?", companyID).Order("is_primary DESC").Find(&out).Error
	return out, err
}
