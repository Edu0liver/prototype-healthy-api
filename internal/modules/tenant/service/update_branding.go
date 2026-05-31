package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/google/uuid"
)

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
