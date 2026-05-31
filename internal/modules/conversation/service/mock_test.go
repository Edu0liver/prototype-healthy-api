package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/events"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// mockRepo is a function-backed conversation Repository.
type mockRepo struct {
	findContactFn      func(ctx context.Context, channelID uuid.UUID, remoteJID string) (*models.Contact, error)
	createContactFn    func(ctx context.Context, c *models.Contact) error
	updateContactFn    func(ctx context.Context, c *models.Contact) error
	findOpenConvFn     func(ctx context.Context, contactID uuid.UUID) (*models.Conversation, error)
	createConvFn       func(ctx context.Context, c *models.Conversation) error
	updateConvFn       func(ctx context.Context, c *models.Conversation) error
	getConvFn          func(ctx context.Context, id uuid.UUID) (*models.Conversation, error)
	listConvFn         func(ctx context.Context, f repository.ConversationFilter) ([]models.Conversation, error)
	insertMessageFn    func(ctx context.Context, m *models.Message) error
	updateStatusFn     func(ctx context.Context, externalID, status string) error
	recentMessagesFn   func(ctx context.Context, convID uuid.UUID, limit int) ([]models.Message, error)
	listMessagesFn     func(ctx context.Context, convID uuid.UUID) ([]models.Message, error)
}

func (m *mockRepo) FindContact(ctx context.Context, channelID uuid.UUID, remoteJID string) (*models.Contact, error) {
	if m.findContactFn != nil {
		return m.findContactFn(ctx, channelID, remoteJID)
	}
	return &models.Contact{}, nil
}
func (m *mockRepo) CreateContact(ctx context.Context, c *models.Contact) error {
	if m.createContactFn != nil {
		return m.createContactFn(ctx, c)
	}
	return nil
}
func (m *mockRepo) UpdateContact(ctx context.Context, c *models.Contact) error {
	if m.updateContactFn != nil {
		return m.updateContactFn(ctx, c)
	}
	return nil
}
func (m *mockRepo) FindOpenConversation(ctx context.Context, contactID uuid.UUID) (*models.Conversation, error) {
	if m.findOpenConvFn != nil {
		return m.findOpenConvFn(ctx, contactID)
	}
	return &models.Conversation{}, nil
}
func (m *mockRepo) CreateConversation(ctx context.Context, c *models.Conversation) error {
	if m.createConvFn != nil {
		return m.createConvFn(ctx, c)
	}
	return nil
}
func (m *mockRepo) UpdateConversation(ctx context.Context, c *models.Conversation) error {
	if m.updateConvFn != nil {
		return m.updateConvFn(ctx, c)
	}
	return nil
}
func (m *mockRepo) GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	if m.getConvFn != nil {
		return m.getConvFn(ctx, id)
	}
	return &models.Conversation{ID: id}, nil
}
func (m *mockRepo) ListConversations(ctx context.Context, f repository.ConversationFilter) ([]models.Conversation, error) {
	if m.listConvFn != nil {
		return m.listConvFn(ctx, f)
	}
	return nil, nil
}
func (m *mockRepo) InsertMessage(ctx context.Context, msg *models.Message) error {
	if m.insertMessageFn != nil {
		return m.insertMessageFn(ctx, msg)
	}
	return nil
}
func (m *mockRepo) UpdateMessageStatusByExternalID(ctx context.Context, externalID, status string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, externalID, status)
	}
	return nil
}
func (m *mockRepo) RecentMessages(ctx context.Context, convID uuid.UUID, limit int) ([]models.Message, error) {
	if m.recentMessagesFn != nil {
		return m.recentMessagesFn(ctx, convID, limit)
	}
	return nil, nil
}
func (m *mockRepo) ListMessages(ctx context.Context, convID uuid.UUID) ([]models.Message, error) {
	if m.listMessagesFn != nil {
		return m.listMessagesFn(ctx, convID)
	}
	return nil, nil
}

// newSvc builds a service with a best-effort publisher pointed at a dead Redis
// address: Publish marshals and tries to send, the send fails and is logged at
// Debug, so unit tests never block or panic on realtime fan-out.
func newSvc(repo Repository) *Service {
	rdb := &redisx.Client{Client: redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})}
	return New(repo, events.New(rdb, zap.NewNop()))
}
