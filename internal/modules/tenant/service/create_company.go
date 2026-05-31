package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
)

// CreateCompany provisions a new tenant with default branding (public signup).
func (s *Service) CreateCompany(ctx context.Context, in dto.CreateCompanyRequest) (*models.Company, error) {
	var created *models.Company
	err := s.db.System(ctx, func(ctx context.Context) error {
		taken, err := s.repo.SlugExists(ctx, in.Slug)
		if err != nil {
			return err
		}
		if taken {
			return ErrSlugTaken
		}

		id := mustUUIDv7()
		plan := in.Plan
		if plan == "" {
			plan = "free"
		}
		company := &models.Company{ID: id, Name: in.Name, Slug: in.Slug, Status: "active", Plan: plan}
		if err := s.repo.CreateCompany(ctx, company); err != nil {
			return err
		}
		if err := s.repo.UpsertBranding(ctx, &models.CompanyBranding{CompanyID: id}); err != nil {
			return err
		}
		created = company
		return nil
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}
