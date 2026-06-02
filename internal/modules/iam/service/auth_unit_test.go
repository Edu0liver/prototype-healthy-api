package service

import (
	"context"
	"testing"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// authSvc builds a Service with only a token manager wired. db/rdb are nil; the
// tested code paths return before touching them.
func authSvc() (*Service, *token.Manager) {
	tok := token.New("unit-test-secret", 15*time.Minute, time.Hour)
	return New(nil, nil, tok, nil, nil, nil), tok
}

func TestRefresh_RejectsGarbageToken(t *testing.T) {
	svc, _ := authSvc()
	_, err := svc.Refresh(context.Background(), "not-a-jwt")
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestRefresh_RejectsAccessTokenAsRefresh(t *testing.T) {
	svc, tok := authSvc()
	access, err := tok.GenerateAccess(uuid.New(), uuid.New(), "admin")
	require.NoError(t, err)
	_, err = svc.Refresh(context.Background(), access) // wrong token type
	require.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestLogout_IsIdempotentForBadTokens(t *testing.T) {
	svc, tok := authSvc()
	access, _ := tok.GenerateAccess(uuid.New(), uuid.New(), "admin")

	// Garbage, wrong-type, and valid refresh tokens all succeed (nothing to leak).
	require.NoError(t, svc.Logout(context.Background(), "garbage"))
	require.NoError(t, svc.Logout(context.Background(), access))

	refresh, err := tok.GenerateRefresh(uuid.New(), uuid.New())
	require.NoError(t, err)
	require.NoError(t, svc.Logout(context.Background(), refresh))
}
