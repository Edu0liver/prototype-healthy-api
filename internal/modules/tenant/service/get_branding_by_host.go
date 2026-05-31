package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/repository"
)

// GetBrandingByHost resolves a Host header to its tenant branding (public,
// no auth — backs GET /branding/host?host=).
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
