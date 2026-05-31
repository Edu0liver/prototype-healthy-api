package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	"github.com/google/uuid"
)

// Me returns the authenticated user (caller's tenant scope).
func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	u, err := s.repo.FindByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrUserNotFound
	}
	return u, err
}
