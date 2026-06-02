package service

import (
	"context"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"github.com/google/uuid"
)

// Refresh validates a refresh token and issues a fresh access/refresh pair,
// reloading the user's current role. The presented refresh token is rotated:
// its jti is denylisted so it cannot be reused (mitigates token theft/replay).
func (s *Service) Refresh(ctx context.Context, refreshToken string) (*Tokens, error) {
	claims, err := s.tokens.Parse(refreshToken)
	if err != nil || claims.Type != token.TypeRefresh {
		return nil, ErrInvalidCredentials
	}
	if s.isRefreshRevoked(ctx, claims.ID) {
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

	// Rotate: revoke the presented refresh token for its remaining lifetime so a
	// stolen copy cannot be replayed after the legitimate client refreshes.
	if claims.ExpiresAt != nil {
		s.revokeRefresh(ctx, claims.ID, time.Until(claims.ExpiresAt.Time))
	}
	return s.issueTokens(companyID, userID, roleName)
}
