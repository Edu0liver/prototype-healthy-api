// Package repositories implements tenant-scoped persistence for channels.
package repositories

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a channel does not exist in the tenant scope.
var ErrNotFound = errors.New("channel: not found")

// Repository persists channels.
type Repository struct{}

// New builds the repository.
func New() *Repository { return &Repository{} }

func wrap(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

// Create inserts a channel.
func (r *Repository) Create(ctx context.Context, c *models.Channel) error {
	return wrap(database.MustTx(ctx).Create(c).Error)
}

// Update saves channel changes within the tenant scope.
func (r *Repository) Update(ctx context.Context, c *models.Channel) error {
	return wrap(database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Save(c).Error)
}

// Get loads a channel by id within the tenant scope.
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	var c models.Channel
	if err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).First(&c, "id = ?", id).Error; err != nil {
		return nil, wrap(err)
	}
	return &c, nil
}

// List returns all channels in the tenant.
func (r *Repository) List(ctx context.Context) ([]models.Channel, error) {
	var out []models.Channel
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Order("created_at DESC").Find(&out).Error
	return out, err
}
