package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/repository"
	"github.com/google/uuid"
)

// GetCompany loads the caller's company (authenticated; uses request tx).
func (s *Service) GetCompany(ctx context.Context, id uuid.UUID) (*models.Company, error) {
	c, err := s.repo.GetCompany(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrCompanyNotFound
	}
	return c, err
}
