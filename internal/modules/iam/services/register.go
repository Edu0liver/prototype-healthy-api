package services

import (
	"context"
	"errors"
	"strings"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repositories"
	"github.com/google/uuid"
)

// RegisterFirstAdmin bootstraps the initial admin for a freshly created company.
// It is only permitted while the company has zero users (self-serve onboarding).
func (s *Service) RegisterFirstAdmin(ctx context.Context, slug, email, password, name string) (*models.User, error) {
	var companyID uuid.UUID
	if err := s.db.System(ctx, func(ctx context.Context) error {
		id, err := s.repo.CompanyIDBySlug(ctx, slug)
		companyID = id
		return err
	}); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	var admin *models.User
	err = s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
		n, err := s.repo.CountUsers(ctx)
		if err != nil {
			return err
		}
		if n > 0 {
			return ErrAdminExists
		}
		admin = &models.User{
			ID:           mustUUIDv7(),
			CompanyID:    companyID,
			Email:        strings.ToLower(email),
			PasswordHash: hash,
			Name:         name,
			Role:         "admin",
			Status:       "active",
		}
		return s.repo.Create(ctx, admin)
	})
	if err != nil {
		return nil, err
	}
	return admin, nil
}
