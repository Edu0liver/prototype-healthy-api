// Package repositories implements tenant-scoped persistence for agents.
package repositories

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound is returned when an agent does not exist in the tenant scope.
var ErrNotFound = errors.New("agent: not found")

// Repository persists agents.
type Repository struct{}

// New builds the repository.
func New() *Repository { return &Repository{} }

func wrap(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

// Create inserts an agent.
func (r *Repository) Create(ctx context.Context, a *models.Agent) error {
	return wrap(database.MustTx(ctx).Create(a).Error)
}

// Update saves agent changes (scoped to the tenant).
func (r *Repository) Update(ctx context.Context, a *models.Agent) error {
	return wrap(database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Save(a).Error)
}

// Get loads an agent by id within the tenant scope.
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*models.Agent, error) {
	var a models.Agent
	if err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).First(&a, "id = ?", id).Error; err != nil {
		return nil, wrap(err)
	}
	return &a, nil
}

// List returns all agents in the tenant.
func (r *Repository) List(ctx context.Context) ([]models.Agent, error) {
	var out []models.Agent
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Order("created_at DESC").Find(&out).Error
	return out, err
}

// Delete removes an agent within the tenant scope.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return wrap(database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Delete(&models.Agent{}, "id = ?", id).Error)
}
