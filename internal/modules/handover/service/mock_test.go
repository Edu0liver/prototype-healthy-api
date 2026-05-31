package service

import (
	"context"

	convmodels "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	convrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	convsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/events"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// mockConvRepo satisfies the conversation service.Repository so handover can run
// against a real convsvc.Service without a database. Only the methods exercised
// by Take/Return/Close (GetConversation, UpdateConversation) are wired.
type mockConvRepo struct {
	getConvFn    func(ctx context.Context, id uuid.UUID) (*convmodels.Conversation, error)
	updateConvFn func(ctx context.Context, c *convmodels.Conversation) error
}

func (m *mockConvRepo) FindContact(context.Context, uuid.UUID, string) (*convmodels.Contact, error) {
	return &convmodels.Contact{}, nil
}
func (m *mockConvRepo) CreateContact(context.Context, *convmodels.Contact) error { return nil }
func (m *mockConvRepo) UpdateContact(context.Context, *convmodels.Contact) error { return nil }
func (m *mockConvRepo) FindOpenConversation(context.Context, uuid.UUID) (*convmodels.Conversation, error) {
	return &convmodels.Conversation{}, nil
}
func (m *mockConvRepo) CreateConversation(context.Context, *convmodels.Conversation) error { return nil }
func (m *mockConvRepo) UpdateConversation(ctx context.Context, c *convmodels.Conversation) error {
	if m.updateConvFn != nil {
		return m.updateConvFn(ctx, c)
	}
	return nil
}
func (m *mockConvRepo) GetConversation(ctx context.Context, id uuid.UUID) (*convmodels.Conversation, error) {
	if m.getConvFn != nil {
		return m.getConvFn(ctx, id)
	}
	return &convmodels.Conversation{ID: id}, nil
}
func (m *mockConvRepo) ListConversations(context.Context, convrepo.ConversationFilter) ([]convmodels.Conversation, error) {
	return nil, nil
}
func (m *mockConvRepo) InsertMessage(context.Context, *convmodels.Message) error            { return nil }
func (m *mockConvRepo) UpdateMessageStatusByExternalID(context.Context, string, string) error { return nil }
func (m *mockConvRepo) RecentMessages(context.Context, uuid.UUID, int) ([]convmodels.Message, error) {
	return nil, nil
}
func (m *mockConvRepo) ListMessages(context.Context, uuid.UUID) ([]convmodels.Message, error) {
	return nil, nil
}

// deadRedis returns a redisx client pointed at an unreachable address; handover
// only uses it for best-effort state mirroring (errors ignored), so commands
// fail fast without panicking.
func deadRedis() *redisx.Client {
	return &redisx.Client{Client: redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})}
}

// newConvSvc builds a conversation service backed by the given mock repo and a
// best-effort publisher at a dead Redis address.
func newConvSvc(repo *mockConvRepo) *convsvc.Service {
	rdb := &redisx.Client{Client: redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})}
	return convsvc.New(repo, events.New(rdb, zap.NewNop()))
}
