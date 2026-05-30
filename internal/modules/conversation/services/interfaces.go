package services

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repositories"
	"github.com/google/uuid"
)

// Repository is the persistence contract consumed by the conversation service.
type Repository interface {
	FindContact(ctx context.Context, channelID uuid.UUID, remoteJID string) (*models.Contact, error)
	CreateContact(ctx context.Context, c *models.Contact) error
	UpdateContact(ctx context.Context, c *models.Contact) error

	FindOpenConversation(ctx context.Context, contactID uuid.UUID) (*models.Conversation, error)
	CreateConversation(ctx context.Context, c *models.Conversation) error
	UpdateConversation(ctx context.Context, c *models.Conversation) error
	GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error)
	ListConversations(ctx context.Context, f repositories.ConversationFilter) ([]models.Conversation, error)

	InsertMessage(ctx context.Context, m *models.Message) error
	UpdateMessageStatusByExternalID(ctx context.Context, externalID, status string) error
	RecentMessages(ctx context.Context, convID uuid.UUID, limit int) ([]models.Message, error)
	ListMessages(ctx context.Context, convID uuid.UUID) ([]models.Message, error)
}
