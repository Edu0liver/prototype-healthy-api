package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"github.com/google/uuid"
)

// Tokens is an issued access/refresh pair.
type Tokens struct {
	Access  string
	Refresh string
}

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

// Refresh validates a refresh token and issues a fresh access token, reloading
// the user's current role.
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	claims, err := s.tokens.Parse(refreshToken)
	if err != nil || claims.Type != token.TypeRefresh {
		return nil, ErrInvalidCredentials
	}
	companyID, err := uuid.Parse(claims.CompanyID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	var roleName string
	if err := s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
		user, err := s.repo.FindByID(ctx, userID)
		if err != nil {
			return err
		}
		if user.Status == "disabled" {
			return ErrUserDisabled
		}
		roleName = user.Role.Name
		return nil
	}); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokens(companyID, userID, roleName)
}

func (s *Service) issueTokens(companyID, userID uuid.UUID, role string) (*Tokens, error) {
	access, err := s.tokens.GenerateAccess(companyID, userID, role)
	if err != nil {
		return nil, err
	}
	refresh, err := s.tokens.GenerateRefresh(companyID, userID)
	if err != nil {
		return nil, err
	}
	return &Tokens{Access: access, Refresh: refresh}, nil
}
