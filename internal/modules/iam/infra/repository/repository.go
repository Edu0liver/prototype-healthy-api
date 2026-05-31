// Package repositories implements persistence for the iam module. The users
// table is under RLS, so reads/writes require a tenant-scoped context
// (database.MustTx). Login resolves company_id via a SECURITY DEFINER function
// that bypasses RLS, then opens the tenant scope for password verification.
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

// FindByEmailGlobal resolves a user by globally-unique email without RLS.
// Uses find_user_by_email() (SECURITY DEFINER) and returns the user's id and company_id.
func (r *Repository) FindByEmailGlobal(ctx context.Context, email string) (userID, companyID uuid.UUID, err error) {
	var row struct {
		ID        uuid.UUID `gorm:"column:id"`
		CompanyID uuid.UUID `gorm:"column:company_id"`
	}
	res := database.MustTx(ctx).Raw("SELECT id, company_id FROM find_user_by_email(?)", email).Scan(&row)
	if res.Error != nil {
		return uuid.Nil, uuid.Nil, wrap(res.Error)
	}
	if row.ID == uuid.Nil {
		return uuid.Nil, uuid.Nil, ErrNotFound
	}
	return row.ID, row.CompanyID, nil
}

// FindRoleByName loads a system role by its name (roles table has no RLS).
func (r *Repository) FindRoleByName(ctx context.Context, name string) (*models.SystemRole, error) {
	var role models.SystemRole
	if err := database.MustTx(ctx).Where("name = ?", name).First(&role).Error; err != nil {
		return nil, wrap(err)
	}
	return &role, nil
}

// Create inserts a user.
func (r *Repository) Create(ctx context.Context, u *models.User) error {
	return wrap(database.MustTx(ctx).Create(u).Error)
}

// Update saves user changes.
func (r *Repository) Update(ctx context.Context, u *models.User) error {
	return wrap(database.MustTx(ctx).Save(u).Error)
}

// FindByEmail loads a user (with role) by email within the current tenant scope.
func (r *Repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Preload("Role").First(&u, "email = ?", email).Error
	if err != nil {
		return nil, wrap(err)
	}
	return &u, nil
}

// FindByID loads a user (with role) by id within the current tenant scope.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Preload("Role").First(&u, "id = ?", id).Error
	if err != nil {
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

// ListByCompany returns all users (with roles) in the current tenant.
func (r *Repository) ListByCompany(ctx context.Context) ([]models.User, error) {
	var out []models.User
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Preload("Role").Order("created_at DESC").Find(&out).Error
	return out, err
}
