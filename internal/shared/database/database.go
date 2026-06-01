// Package database wires GORM over PostgreSQL and enforces the tenant RLS
// session variable. Every tenant-scoped query runs inside a transaction that
// issues `SET LOCAL app.current_company_id`, so Row-Level Security filters rows
// for exactly one company and the setting never leaks across pooled connections.
package database

import (
	"context"
	"fmt"
	stdlog "log"
	"os"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB wraps the root *gorm.DB.
type DB struct {
	*gorm.DB
}

type ctxKey struct{}

// New opens the connection pool and registers graceful shutdown.
func New(cfg *config.Config, lc fx.Lifecycle, log *zap.Logger) (*DB, error) {
	gormLogLevel := logger.Warn
	if cfg.IsProduction() {
		gormLogLevel = logger.Error
	}

	// IgnoreRecordNotFoundError: "record not found" is an expected outcome for
	// lookups like Host->tenant resolution (e.g. an unmapped domain). Logging it
	// as an error floods the log on every miss; it is handled in the app layer.
	gormLogger := logger.New(
		stdlog.New(os.Stdout, "\r\n", stdlog.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormLogLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	gdb, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}

	sqlDB, err := gdb.DB()
	if err != nil {
		return nil, fmt.Errorf("database: pool handle: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return sqlDB.PingContext(ctx) },
		OnStop: func(context.Context) error {
			log.Info("closing database pool")
			return sqlDB.Close()
		},
	})

	return &DB{gdb}, nil
}

// Tenant runs fn inside a transaction scoped to companyID. It sets the RLS
// session variable with SET LOCAL (transaction lifetime) and injects the
// transaction handle into the context for repositories to pick up.
func (d *DB) Tenant(ctx context.Context, companyID uuid.UUID, fn func(ctx context.Context) error) error {
	ctx = ensureIdentity(ctx, companyID)
	return d.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := setTenant(tx, companyID); err != nil {
			return err
		}
		return fn(WithTx(ctx, tx))
	})
}

// ensureIdentity guarantees the context carries the tenant id so the app-layer
// TenantScope filter works in programmatic (worker) paths as well as HTTP.
func ensureIdentity(ctx context.Context, companyID uuid.UUID) context.Context {
	if appctx.CompanyID(ctx) == uuid.Nil {
		return appctx.With(ctx, appctx.Identity{CompanyID: companyID})
	}
	return ctx
}

// TenantScope is a GORM scope that enforces company_id at the application layer
// (PRD invariant 1) on top of RLS. Use on every tenant-scoped read.
func TenantScope(ctx context.Context) func(*gorm.DB) *gorm.DB {
	companyID := appctx.CompanyID(ctx)
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("company_id = ?", companyID)
	}
}

// System runs fn against the root pool without tenant scoping. Use only for
// platform operations (the tenant registry, webhook intake before resolution).
func (d *DB) System(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(WithTx(ctx, d.WithContext(ctx)))
}

func setTenant(tx *gorm.DB, companyID uuid.UUID) error {
	// SET LOCAL value cannot be parameterized; companyID is a validated UUID.
	if err := tx.Exec(fmt.Sprintf("SET LOCAL app.current_company_id = '%s'", companyID.String())).Error; err != nil {
		return fmt.Errorf("database: set tenant: %w", err)
	}
	return nil
}

// WithTx stores a query handle (tx or session) in the context.
func WithTx(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxKey{}, db)
}

// TxFromContext returns the request/tenant-scoped handle if present.
func TxFromContext(ctx context.Context) (*gorm.DB, bool) {
	db, ok := ctx.Value(ctxKey{}).(*gorm.DB)
	return db, ok
}

// MustTx returns the scoped handle or panics — repositories require a scope.
func MustTx(ctx context.Context) *gorm.DB {
	db, ok := TxFromContext(ctx)
	if !ok {
		panic("database: no tenant/tx scope in context (missing TenantResolver middleware or DB.Tenant wrapper)")
	}
	return db.WithContext(ctx)
}

// BeginTenant manually starts a tenant-scoped transaction for callers (such as
// HTTP middleware) that cannot use the closure form. Caller MUST Commit/Rollback.
func (d *DB) BeginTenant(ctx context.Context, companyID uuid.UUID) (*gorm.DB, error) {
	tx := d.WithContext(ensureIdentity(ctx, companyID)).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	if err := setTenant(tx, companyID); err != nil {
		tx.Rollback()
		return nil, err
	}
	return tx, nil
}
