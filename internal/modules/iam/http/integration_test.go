package http_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/testsupport"
	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type noopMailer struct{}

func (noopMailer) Send(string, string, string) error { return nil }

func newService(t *testing.T, db *database.DB) *service.Service {
	cfg := &config.Config{}
	cfg.App.PublicBaseURL = "http://test"
	cfg.JWT.Secret = "test-secret-please-change"
	cfg.JWT.AccessTTL = 15 * time.Minute
	cfg.JWT.RefreshTTL = time.Hour
	return service.New(repository.New(), db, token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL), noopMailer{}, cfg)
}

func seedCompany(t *testing.T, db *database.DB, slug string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	err := db.System(context.Background(), func(ctx context.Context) error {
		return database.MustTx(ctx).Exec(
			"INSERT INTO companies (id, name, slug, status, plan) VALUES (?, ?, ?, 'active', 'free')",
			id, slug, slug,
		).Error
	})
	require.NoError(t, err)
	return id
}

// TestTenantIsolation is the mandatory tenant-leakage test (PRD invariant 1,
// RF-WL-03): a user of company A must never see company B's data, even though
// both rows live in the same table.
func TestTenantIsolation(t *testing.T) {
	db := testsupport.NewPostgres(t)
	svc := newService(t, db)
	ctx := context.Background()

	slugA := "company-a"
	slugB := "company-b"
	seedCompany(t, db, slugA)
	seedCompany(t, db, slugB)

	_, err := svc.RegisterFirstAdmin(ctx, slugA, "admin@a.com", "supersecret", "A")
	require.NoError(t, err)
	_, err = svc.RegisterFirstAdmin(ctx, slugB, "admin@b.com", "supersecret", "B")
	require.NoError(t, err)

	// A's admin can log in; the scope is resolved by company slug.
	tokensA, userA, err := svc.Login(ctx, slugA, "admin@a.com", "supersecret")
	require.NoError(t, err)
	require.NotEmpty(t, tokensA.Access)

	// Listing users inside A's tenant scope returns only A's users.
	err = db.Tenant(ctx, userA.CompanyID, func(ctx context.Context) error {
		users, err := svc.ListUsers(ctx)
		require.NoError(t, err)
		require.Len(t, users, 1)
		require.Equal(t, "admin@a.com", users[0].Email)
		return nil
	})
	require.NoError(t, err)

	// A user from B is invisible inside A's scope (email is unique per company).
	err = db.Tenant(ctx, userA.CompanyID, func(ctx context.Context) error {
		repo := repository.New()
		_, err := repo.FindByEmail(ctx, "admin@b.com")
		require.True(t, errors.Is(err, repository.ErrNotFound), "expected B's user invisible to A, got %v", err)
		return nil
	})
	require.NoError(t, err)

	// Cross-tenant login is rejected: A's credentials under B's slug fail.
	_, _, err = svc.Login(ctx, slugB, "admin@a.com", "supersecret")
	require.Error(t, err)
}
