// Package testsupport provides shared helpers for integration tests: a real
// PostgreSQL (pgvector) container with all migrations applied.
package testsupport

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
)

// SeedActiveSubscription gives a company an active subscription on the starter
// plan so the subscription gate (middleware.RequireActiveSubscription) lets the
// operational modules through in integration tests.
func SeedActiveSubscription(t *testing.T, db *database.DB, companyID uuid.UUID) {
	t.Helper()
	err := db.System(context.Background(), func(ctx context.Context) error {
		return database.MustTx(ctx).Exec(
			`INSERT INTO subscriptions
			   (id, company_id, plan_id, status, current_period_start, current_period_end)
			 SELECT gen_random_uuid(), ?, id, 'active', now(), now() + interval '1 month'
			   FROM plans WHERE code = 'starter'
			 ON CONFLICT (company_id) DO UPDATE SET
			   status = 'active', current_period_end = now() + interval '1 month'`,
			companyID,
		).Error
	})
	if err != nil {
		t.Fatalf("seed active subscription: %v", err)
	}
}

// repoRoot returns the repository root relative to this source file.
func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	// internal/shared/testsupport/postgres.go -> up 3 dirs to repo root.
	return filepath.Join(filepath.Dir(file), "..", "..", "..")
}

// NewPostgres starts a pgvector container, applies all migrations, and returns
// a connected *database.DB. The container is terminated via t.Cleanup.
func NewPostgres(t *testing.T) *database.DB {
	t.Helper()
	ctx := context.Background()

	container, err := postgres.Run(ctx, "pgvector/pgvector:pg17",
		postgres.WithDatabase("lumia"),
		postgres.WithUsername("lumia"),
		postgres.WithPassword("lumia"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("connection string: %v", err)
	}

	// Apply migrations with database/sql + goose.
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open sql: %v", err)
	}
	defer sqlDB.Close()
	goose.SetDialect("postgres")
	if err := goose.Up(sqlDB, filepath.Join(repoRoot(), "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	gdb, err := gorm.Open(gormpg.Open(dsn), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		t.Fatalf("open gorm: %v", err)
	}
	return &database.DB{DB: gdb}
}
