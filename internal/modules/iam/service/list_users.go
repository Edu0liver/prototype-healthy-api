package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
)

// ListUsers returns all users in the caller's tenant.
func (s *Service) ListUsers(ctx context.Context) ([]models.User, error) {
	return s.repo.ListByCompany(ctx)
}
