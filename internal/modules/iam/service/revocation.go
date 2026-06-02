package service

import (
	"context"
	"time"
)

// refreshDenyPrefix namespaces revoked refresh-token jti entries in Redis. Each
// entry expires when the underlying token would have expired anyway, so the
// denylist stays bounded.
const refreshDenyPrefix = "auth:refresh:deny:"

// isRefreshRevoked reports whether a refresh token's jti has been denylisted.
// Fails open (false) when Redis is unavailable.
func (s *Service) isRefreshRevoked(ctx context.Context, jti string) bool {
	if s.rdb == nil || jti == "" {
		return false
	}
	n, err := s.rdb.Exists(ctx, refreshDenyPrefix+jti).Result()
	if err != nil {
		return false
	}
	return n > 0
}

// revokeRefresh denylists a refresh-token jti for the remaining token lifetime.
func (s *Service) revokeRefresh(ctx context.Context, jti string, ttl time.Duration) {
	if s.rdb == nil || jti == "" || ttl <= 0 {
		return
	}
	s.rdb.Set(ctx, refreshDenyPrefix+jti, "1", ttl)
}
