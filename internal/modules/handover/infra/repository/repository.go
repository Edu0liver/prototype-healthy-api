// Package repositories provides handover dispatch lookups (channel creds +
// recipient) for sending operator replies.
package repository

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound indicates the conversation/channel was not found in scope.
var ErrNotFound = errors.New("handover: not found")

// Repository reads dispatch info.
type Repository struct{}

// New builds the repository.
func New() *Repository { return &Repository{} }

// DispatchInfo is what we need to send an outbound message on a conversation.
type DispatchInfo struct {
	Instance    string `gorm:"column:instance"`
	APIKeyEnc   string `gorm:"column:apikey_enc"`
	ChannelType string `gorm:"column:channel_type"`
	RemoteJID   string `gorm:"column:remote_jid"`
}

// LoadDispatchInfo joins conversation→channel→contact (tenant-scoped).
func (r *Repository) LoadDispatchInfo(ctx context.Context, convID uuid.UUID) (*DispatchInfo, error) {
	var info DispatchInfo
	err := database.MustTx(ctx).
		Table("conversations AS cv").
		Select("ch.evolution_instance_name AS instance, ch.evolution_apikey_enc AS apikey_enc, ch.type AS channel_type, ct.remote_jid AS remote_jid").
		Joins("JOIN channels ch ON ch.id = cv.channel_id").
		Joins("JOIN contacts ct ON ct.id = cv.contact_id").
		Where("cv.company_id = ? AND cv.id = ?", appctx.CompanyID(ctx), convID).
		Take(&info).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &info, err
}
