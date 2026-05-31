package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/google/uuid"
)

// ListDomains returns the caller tenant's domains.
func (s *Service) ListDomains(ctx context.Context, companyID uuid.UUID) ([]models.CompanyDomain, error) {
	return s.repo.ListDomains(ctx, companyID)
}
