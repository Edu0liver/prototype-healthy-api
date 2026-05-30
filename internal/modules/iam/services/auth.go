package services

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/token"
	"github.com/google/uuid"
)

// Tokens is an issued access/refresh pair.
type Tokens struct {
	Access  string
	Refresh string
}

// Login authenticates a user within a company (resolved by slug) and issues tokens.
func (s *Service) Login(ctx context.Context, slug, email, password string) (*Tokens, *models.User, error) {
	var companyID uuid.UUID
	if err := s.db.System(ctx, func(ctx context.Context) error {
		id, err := s.repo.CompanyIDBySlug(ctx, slug)
		companyID = id
		return err
	}); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	var user *models.User
	if err := s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
		u, err := s.repo.FindByEmail(ctx, email)
		user = u
		return err
	}); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	if user.Status == "disabled" {
		return nil, nil, ErrUserDisabled
	}
	ok, err := verifyPassword(password, user.PasswordHash)
	if err != nil || !ok {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := s.issueTokens(companyID, user.ID, user.Role)
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

	var user *models.User
	if err := s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
		u, err := s.repo.FindByID(ctx, userID)
		user = u
		return err
	}); err != nil {
		return nil, ErrInvalidCredentials
	}
	if user.Status == "disabled" {
		return nil, ErrUserDisabled
	}
	return s.issueTokens(companyID, user.ID, user.Role)
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
