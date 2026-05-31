package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	"github.com/google/uuid"
)

// Login authenticates a user by globally-unique email and issues tokens.
func (s *Service) Login(ctx context.Context, email, password string) (*Tokens, *models.User, error) {
	var userID, companyID uuid.UUID
	if err := s.db.System(ctx, func(ctx context.Context) error {
		uid, cid, err := s.repo.FindByEmailGlobal(ctx, email)
		userID, companyID = uid, cid
		return err
	}); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	var user *models.User
	if err := s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
		u, err := s.repo.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		if u.Status == "disabled" {
			return ErrUserDisabled
		}
		ok, err := verifyPassword(password, u.PasswordHash)
		if err != nil || !ok {
			return ErrInvalidCredentials
		}
		user = u
		return nil
	}); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	tokens, err := s.issueTokens(companyID, userID, user.Role.Name)
	if err != nil {
		return nil, nil, err
	}
	return tokens, user, nil
}
