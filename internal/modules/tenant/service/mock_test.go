package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/google/uuid"
)

// mockRepo is a function-backed Repository for the service unit tests. Use cases
// wrapped in db.System (CreateCompany, GetBrandingByHost) require a real DB and
// are exercised by the http integration test instead.
type mockRepo struct {
	createCompanyFn       func(ctx context.Context, c *models.Company) error
	getCompanyFn          func(ctx context.Context, id uuid.UUID) (*models.Company, error)
	findCompanyByDomainFn func(ctx context.Context, domain string) (*models.Company, error)
	slugExistsFn          func(ctx context.Context, slug string) (bool, error)
	upsertBrandingFn      func(ctx context.Context, b *models.CompanyBranding) error
	getBrandingFn         func(ctx context.Context, companyID uuid.UUID) (*models.CompanyBranding, error)
	addDomainFn           func(ctx context.Context, d *models.CompanyDomain) error
	listDomainsFn         func(ctx context.Context, companyID uuid.UUID) ([]models.CompanyDomain, error)
}

func (m *mockRepo) CreateCompany(ctx context.Context, c *models.Company) error {
	if m.createCompanyFn != nil {
		return m.createCompanyFn(ctx, c)
	}
	return nil
}

func (m *mockRepo) GetCompany(ctx context.Context, id uuid.UUID) (*models.Company, error) {
	if m.getCompanyFn != nil {
		return m.getCompanyFn(ctx, id)
	}
	return &models.Company{ID: id}, nil
}

func (m *mockRepo) FindCompanyByDomain(ctx context.Context, domain string) (*models.Company, error) {
	if m.findCompanyByDomainFn != nil {
		return m.findCompanyByDomainFn(ctx, domain)
	}
	return &models.Company{}, nil
}

func (m *mockRepo) SlugExists(ctx context.Context, slug string) (bool, error) {
	if m.slugExistsFn != nil {
		return m.slugExistsFn(ctx, slug)
	}
	return false, nil
}

func (m *mockRepo) UpsertBranding(ctx context.Context, b *models.CompanyBranding) error {
	if m.upsertBrandingFn != nil {
		return m.upsertBrandingFn(ctx, b)
	}
	return nil
}

func (m *mockRepo) GetBranding(ctx context.Context, companyID uuid.UUID) (*models.CompanyBranding, error) {
	if m.getBrandingFn != nil {
		return m.getBrandingFn(ctx, companyID)
	}
	return &models.CompanyBranding{CompanyID: companyID}, nil
}

func (m *mockRepo) AddDomain(ctx context.Context, d *models.CompanyDomain) error {
	if m.addDomainFn != nil {
		return m.addDomainFn(ctx, d)
	}
	return nil
}

func (m *mockRepo) ListDomains(ctx context.Context, companyID uuid.UUID) ([]models.CompanyDomain, error) {
	if m.listDomainsFn != nil {
		return m.listDomainsFn(ctx, companyID)
	}
	return nil, nil
}
