package service

import (
	"context"
	"strings"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
)

// RegisterFirstAdmin bootstraps the initial admin for a freshly created company.
// It is only permitted while the company has zero users (self-serve onboarding).
func (s *Service) RegisterFirstAdmin(ctx context.Context, companyID uuid.UUID, email, password, name string) (*models.User, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	var exists bool
	if err = s.db.System(ctx, func(ctx context.Context) error {
		return database.MustTx(ctx).Raw("SELECT EXISTS(SELECT 1 FROM companies WHERE id = ?)", companyID).Scan(&exists).Error
	}); err != nil || !exists {
		return nil, ErrCompanyNotFound
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
		role, err := s.repo.FindRoleByName(ctx, "admin")
		if err != nil {
			return err
		}
		admin = &models.User{
			ID:           mustUUIDv7(),
			CompanyID:    companyID,
			Email:        strings.ToLower(email),
			PasswordHash: hash,
			Name:         name,
			RoleID:       role.ID,
			Status:       "active",
		}
		return s.repo.Create(ctx, admin)
	})
	if err != nil {
		return nil, err
	}
	return admin, nil
}
