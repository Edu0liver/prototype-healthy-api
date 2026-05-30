// Package repositories implements tenant-scoped persistence for automations,
// including validation that referenced channel/agent belong to the tenant and
// reflection of the active agent onto the channel row.
package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound indicates the automation does not exist in the tenant scope.
var ErrNotFound = errors.New("automation: not found")

// ErrActiveExists indicates another active automation already exists for the channel.
var ErrActiveExists = errors.New("automation: active automation already exists for channel")

// Repository persists automations.
type Repository struct{}

// New builds the repository.
func New() *Repository { return &Repository{} }

func wrap(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil && strings.Contains(err.Error(), "uniq_active_automation_per_channel") {
		return ErrActiveExists
	}
	return err
}

// Create inserts an automation.
func (r *Repository) Create(ctx context.Context, a *models.Automation) error {
	return wrap(database.MustTx(ctx).Create(a).Error)
}

// Update saves automation changes within the tenant scope.
func (r *Repository) Update(ctx context.Context, a *models.Automation) error {
	return wrap(database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Save(a).Error)
}

// Get loads an automation by id within the tenant scope.
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*models.Automation, error) {
	var a models.Automation
	if err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).First(&a, "id = ?", id).Error; err != nil {
		return nil, wrap(err)
	}
	return &a, nil
}

// List returns all automations in the tenant.
func (r *Repository) List(ctx context.Context) ([]models.Automation, error) {
	var out []models.Automation
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Order("created_at DESC").Find(&out).Error
	return out, err
}

// Delete removes an automation within the tenant scope.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return wrap(database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Delete(&models.Automation{}, "id = ?", id).Error)
}

// ChannelBelongsToTenant reports whether the channel exists in the tenant scope.
func (r *Repository) ChannelBelongsToTenant(ctx context.Context, channelID uuid.UUID) (bool, error) {
	return r.existsScoped(ctx, "channels", channelID)
}

// AgentBelongsToTenant reports whether the agent exists in the tenant scope.
func (r *Repository) AgentBelongsToTenant(ctx context.Context, agentID uuid.UUID) (bool, error) {
	return r.existsScoped(ctx, "agents", agentID)
}

func (r *Repository) existsScoped(ctx context.Context, table string, id uuid.UUID) (bool, error) {
	var n int64
	err := database.MustTx(ctx).Table(table).Scopes(database.TenantScope(ctx)).Where("id = ?", id).Count(&n).Error
	return n > 0, err
}

// SetChannelActiveAgent reflects the active agent onto the channel (nil clears).
func (r *Repository) SetChannelActiveAgent(ctx context.Context, channelID uuid.UUID, agentID *uuid.UUID) error {
	return database.MustTx(ctx).Table("channels").Scopes(database.TenantScope(ctx)).
		Where("id = ?", channelID).Update("active_agent_id", agentID).Error
}
