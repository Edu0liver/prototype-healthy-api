package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/repository"
	"github.com/google/uuid"
)

// GetBranding returns the caller tenant's branding (authenticated).
func (s *Service) GetBranding(ctx context.Context, companyID uuid.UUID) (*models.CompanyBranding, error) {
	b, err := s.repo.GetBranding(ctx, companyID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrBrandingNotFound
	}
	return b, err
}
