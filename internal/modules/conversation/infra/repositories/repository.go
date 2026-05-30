// Package repositories implements tenant-scoped persistence for contacts,
// conversations and messages.
package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound indicates the row does not exist in the tenant scope.
var ErrNotFound = errors.New("conversation: not found")

// ErrDuplicate indicates a message with the same external id already exists.
var ErrDuplicate = errors.New("conversation: duplicate message")

// Repository persists conversation aggregates.
type Repository struct{}

// New builds the repository.
func New() *Repository { return &Repository{} }

func wrap(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

// FindContact returns a contact by channel + remote jid, or ErrNotFound.
func (r *Repository) FindContact(ctx context.Context, channelID uuid.UUID, remoteJID string) (*models.Contact, error) {
	var c models.Contact
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).
		First(&c, "channel_id = ? AND remote_jid = ?", channelID, remoteJID).Error
	if err != nil {
		return nil, wrap(err)
	}
	return &c, nil
}

// CreateContact inserts a contact.
func (r *Repository) CreateContact(ctx context.Context, c *models.Contact) error {
	return wrap(database.MustTx(ctx).Create(c).Error)
}

// UpdateContact saves contact changes.
func (r *Repository) UpdateContact(ctx context.Context, c *models.Contact) error {
	return wrap(database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Save(c).Error)
}

// FindOpenConversation returns the most recent non-closed conversation for a contact.
func (r *Repository) FindOpenConversation(ctx context.Context, contactID uuid.UUID) (*models.Conversation, error) {
	var c models.Conversation
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).
		Where("contact_id = ? AND state <> ?", contactID, "closed").
		Order("opened_at DESC").First(&c).Error
	if err != nil {
		return nil, wrap(err)
	}
	return &c, nil
}

// CreateConversation inserts a conversation.
func (r *Repository) CreateConversation(ctx context.Context, c *models.Conversation) error {
	return wrap(database.MustTx(ctx).Create(c).Error)
}

// UpdateConversation saves conversation changes.
func (r *Repository) UpdateConversation(ctx context.Context, c *models.Conversation) error {
	return wrap(database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Save(c).Error)
}

// GetConversation loads a conversation by id.
func (r *Repository) GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	var c models.Conversation
	if err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).First(&c, "id = ?", id).Error; err != nil {
		return nil, wrap(err)
	}
	return &c, nil
}

// ConversationFilter narrows a conversation listing.
type ConversationFilter struct {
	ChannelID *uuid.UUID
	State     string
	Since     *time.Time
	Limit     int
}

// ListConversations returns conversations matching the filter (tenant-scoped).
func (r *Repository) ListConversations(ctx context.Context, f ConversationFilter) ([]models.Conversation, error) {
	q := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Model(&models.Conversation{})
	if f.ChannelID != nil {
		q = q.Where("channel_id = ?", *f.ChannelID)
	}
	if f.State != "" {
		q = q.Where("state = ?", f.State)
	}
	if f.Since != nil {
		q = q.Where("last_message_at >= ?", *f.Since)
	}
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var out []models.Conversation
	err := q.Order("last_message_at DESC NULLS LAST").Limit(limit).Find(&out).Error
	return out, err
}

// InsertMessage inserts a message, treating the external-id unique violation as
// idempotent (ErrDuplicate).
func (r *Repository) InsertMessage(ctx context.Context, m *models.Message) error {
	err := database.MustTx(ctx).Create(m).Error
	if err != nil && strings.Contains(err.Error(), "uniq_messages_company_external") {
		return ErrDuplicate
	}
	return err
}

// UpdateMessageStatusByExternalID sets delivery status from Evolution updates.
func (r *Repository) UpdateMessageStatusByExternalID(ctx context.Context, externalID, status string) error {
	return database.MustTx(ctx).Model(&models.Message{}).Scopes(database.TenantScope(ctx)).
		Where("external_message_id = ?", externalID).Update("status", status).Error
}

// RecentMessages returns the last N messages of a conversation, oldest-first.
func (r *Repository) RecentMessages(ctx context.Context, convID uuid.UUID, limit int) ([]models.Message, error) {
	var desc []models.Message
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).
		Where("conversation_id = ?", convID).Order("created_at DESC").Limit(limit).Find(&desc).Error
	if err != nil {
		return nil, err
	}
	// reverse to chronological order
	for i, j := 0, len(desc)-1; i < j; i, j = i+1, j-1 {
		desc[i], desc[j] = desc[j], desc[i]
	}
	return desc, nil
}

// ListMessages returns all messages of a conversation, chronological.
func (r *Repository) ListMessages(ctx context.Context, convID uuid.UUID) ([]models.Message, error) {
	var out []models.Message
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).
		Where("conversation_id = ?", convID).Order("created_at ASC").Find(&out).Error
	return out, err
}
