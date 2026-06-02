package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	"github.com/google/uuid"
)

// Login authenticates a user by globally-unique email and issues tokens. It
// enforces a per-email lockout after repeated failures and pays the password
// hashing cost even for unknown accounts to avoid leaking which emails exist.
func (s *Service) Login(ctx context.Context, email, password string) (*Tokens, *models.User, error) {
	if s.loginLocked(ctx, email) {
		return nil, nil, ErrTooManyAttempts
	}

	var userID, companyID uuid.UUID
	if err := s.db.System(ctx, func(ctx context.Context) error {
		uid, cid, err := s.repo.FindByEmailGlobal(ctx, email)
		userID, companyID = uid, cid
		return err
	}); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			dummyVerify(password) // equalize timing vs the found-user path
			s.recordLoginFailure(ctx, email)
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
		if errors.Is(err, repository.ErrNotFound) || errors.Is(err, ErrInvalidCredentials) {
			s.recordLoginFailure(ctx, email)
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	s.resetLoginFailures(ctx, email)
	tokens, err := s.issueTokens(companyID, userID, user.Role.Name)
	if err != nil {
		return nil, nil, err
	}
	return tokens, user, nil
}
