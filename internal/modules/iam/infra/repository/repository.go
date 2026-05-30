// Package repositories implements persistence for the iam module. The users
// table is under RLS, so reads/writes require a tenant-scoped context
// (database.MustTx); login resolves the company id first, then opens that scope.
package repository

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a row does not exist.
var ErrNotFound = errors.New("iam: not found")

// Repository persists users.
type Repository struct{}

// New builds the repository.
func New() *Repository { return &Repository{} }

func wrap(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

// CompanyIDBySlug resolves a company id from its slug (registry, no RLS).
func (r *Repository) CompanyIDBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	var row struct{ ID uuid.UUID }
	err := database.MustTx(ctx).Table("companies").Select("id").Where("slug = ?", slug).Take(&row).Error
	if err != nil {
		return uuid.Nil, wrap(err)
	}
	return row.ID, nil
}

// Create inserts a user.
func (r *Repository) Create(ctx context.Context, u *models.User) error {
	return wrap(database.MustTx(ctx).Create(u).Error)
}

// Update saves user changes.
func (r *Repository) Update(ctx context.Context, u *models.User) error {
	return wrap(database.MustTx(ctx).Save(u).Error)
}

// FindByEmail loads a user by email within the current tenant scope.
func (r *Repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	if err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).First(&u, "email = ?", email).Error; err != nil {
		return nil, wrap(err)
	}
	return &u, nil
}

// FindByID loads a user by id within the current tenant scope.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	if err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).First(&u, "id = ?", id).Error; err != nil {
		return nil, wrap(err)
	}
	return &u, nil
}

// CountUsers returns the number of users in the current tenant scope.
func (r *Repository) CountUsers(ctx context.Context) (int64, error) {
	var n int64
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Model(&models.User{}).Count(&n).Error
	return n, err
}

// ListByCompany returns all users in the current tenant.
func (r *Repository) ListByCompany(ctx context.Context) ([]models.User, error) {
	var out []models.User
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Order("created_at DESC").Find(&out).Error
	return out, err
}
