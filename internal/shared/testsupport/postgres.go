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

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
)

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
