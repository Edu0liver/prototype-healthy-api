package service

import (
	"context"
	"strings"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/google/uuid"
)

// AddDomain registers a custom domain mapping for the caller tenant.
func (s *Service) AddDomain(ctx context.Context, companyID uuid.UUID, in dto.AddDomainRequest) (*models.CompanyDomain, error) {
	d := &models.CompanyDomain{
		ID:        mustUUIDv7(),
		CompanyID: companyID,
		Domain:    strings.ToLower(in.Domain),
		IsPrimary: in.IsPrimary,
	}
	if err := s.repo.AddDomain(ctx, d); err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDomainTaken
		}
		return nil, err
	}
	return d, nil
}
