package service

import (
	"context"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
)

// Logout revokes a refresh token so it can no longer mint new access tokens.
// It is idempotent and intentionally lenient: a malformed or already-expired
// token is treated as a successful logout (nothing left to revoke).
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.tokens.Parse(refreshToken)
	if err != nil || claims.Type != token.TypeRefresh {
		return nil
	}
	var ttl time.Duration
	if claims.ExpiresAt != nil {
		ttl = time.Until(claims.ExpiresAt.Time)
	}
	s.revokeRefresh(ctx, claims.ID, ttl)
	return nil
}
