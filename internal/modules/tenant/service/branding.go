package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/repository"
	"github.com/google/uuid"
)

// GetBrandingByHost resolves a Host header to its tenant branding (public,
// no auth — backs GET /branding?host=).
func (s *Service) GetBrandingByHost(ctx context.Context, host string) (*models.CompanyBranding, error) {
	var out *models.CompanyBranding
	err := s.db.System(ctx, func(ctx context.Context) error {
		company, err := s.repo.FindCompanyByDomain(ctx, host)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrCompanyNotFound
			}
			return err
		}
		b, err := s.repo.GetBranding(ctx, company.ID)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrBrandingNotFound
			}
			return err
		}
		out = b
		return nil
	})
	return out, err
}

// GetBranding returns the caller tenant's branding (authenticated).
func (s *Service) GetBranding(ctx context.Context, companyID uuid.UUID) (*models.CompanyBranding, error) {
	b, err := s.repo.GetBranding(ctx, companyID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrBrandingNotFound
	}
	return b, err
}

// UpdateBranding upserts the caller tenant's white-label settings.
func (s *Service) UpdateBranding(ctx context.Context, companyID uuid.UUID, in dto.UpdateBrandingRequest) (*models.CompanyBranding, error) {
	b := &models.CompanyBranding{
		CompanyID:       companyID,
		LogoURL:         in.LogoURL,
		FaviconURL:      in.FaviconURL,
		PrimaryColor:    in.PrimaryColor,
		SecondaryColor:  in.SecondaryColor,
		EmailSenderName: in.EmailSenderName,
	}
	if err := s.repo.UpsertBranding(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}
