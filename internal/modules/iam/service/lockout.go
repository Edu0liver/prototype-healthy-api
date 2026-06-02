package service

import (
	"context"
	"strings"
)

// loginFailPrefix namespaces the per-email failed-login counters in Redis.
const loginFailPrefix = "login:fail:"

// loginKey normalizes the email so case variations share one counter.
func loginKey(email string) string {
	return loginFailPrefix + strings.ToLower(strings.TrimSpace(email))
}

// loginLocked reports whether the account has exceeded the failed-attempt
// threshold within the lockout window. It fails open (returns false) when Redis
// or config is unavailable so an outage cannot lock everyone out.
func (s *Service) loginLocked(ctx context.Context, email string) bool {
	if s.rdb == nil || s.cfg == nil {
		return false
	}
	max := s.cfg.Security.LoginMaxAttempts
	if max <= 0 {
		return false
	}
	n, err := s.rdb.Get(ctx, loginKey(email)).Int()
	if err != nil {
		return false
	}
	return n >= max
}

// recordLoginFailure increments the failed-attempt counter and (re)arms its TTL.
func (s *Service) recordLoginFailure(ctx context.Context, email string) {
	if s.rdb == nil || s.cfg == nil || s.cfg.Security.LoginMaxAttempts <= 0 {
		return
	}
	key := loginKey(email)
	n, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return
	}
	if n == 1 {
		s.rdb.Expire(ctx, key, s.cfg.Security.LoginLockout)
	}
}

// resetLoginFailures clears the counter after a successful login.
func (s *Service) resetLoginFailures(ctx context.Context, email string) {
	if s.rdb == nil {
		return
	}
	s.rdb.Del(ctx, loginKey(email))
}
