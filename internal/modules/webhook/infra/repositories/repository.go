// Package repositories implements webhook routing/audit persistence.
package repositories

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/webhook/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/google/uuid"
)

// Repository persists webhook events and resolves channel routing.
type Repository struct{}

// New builds the repository.
func New() *Repository { return &Repository{} }

// Resolved is the result of instance->tenant routing.
type Resolved struct {
	CompanyID uuid.UUID
	ChannelID uuid.UUID
}

// ResolveChannel maps an Evolution instance to its company/channel via the
// SECURITY DEFINER function (bypasses RLS for this narrow read). Must run in a
// system (non-tenant) scope. Returns found=false when no channel matches.
func (r *Repository) ResolveChannel(ctx context.Context, instance string) (Resolved, bool, error) {
	var row Resolved
	err := database.MustTx(ctx).Raw(
		`SELECT company_id, channel_id FROM resolve_channel_by_instance(?)`, instance,
	).Scan(&row).Error
	if err != nil {
		return Resolved{}, false, err
	}
	if row.CompanyID == uuid.Nil {
		return Resolved{}, false, nil
	}
	return row, true, nil
}

// InsertEvent records a webhook event for audit (webhook_events is not under RLS).
func (r *Repository) InsertEvent(ctx context.Context, e *models.WebhookEvent) error {
	return database.MustTx(ctx).Create(e).Error
}

// UpdateChannelStatus syncs channels.status (runs in a tenant scope).
func (r *Repository) UpdateChannelStatus(ctx context.Context, channelID uuid.UUID, status string) error {
	return database.MustTx(ctx).Table("channels").Scopes(database.TenantScope(ctx)).
		Where("id = ?", channelID).Update("status", status).Error
}
