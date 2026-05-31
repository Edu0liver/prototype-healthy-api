package service

import (
	"context"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func pipeCtx() context.Context {
	return appctx.With(context.Background(), appctx.Identity{CompanyID: uuid.New()})
}

func TestEnsureContact_ExistingRefreshesPushName(t *testing.T) {
	existing := &models.Contact{ID: uuid.New(), PushName: "Old"}
	updated := false
	svc := newSvc(&mockRepo{
		findContactFn:   func(context.Context, uuid.UUID, string) (*models.Contact, error) { return existing, nil },
		updateContactFn: func(_ context.Context, c *models.Contact) error { updated = true; return nil },
	})
	out, err := svc.EnsureContact(context.Background(), uuid.New(), "jid", "New")
	require.NoError(t, err)
	require.Equal(t, "New", out.PushName)
	require.True(t, updated)
}

func TestEnsureContact_CreatesWhenMissing(t *testing.T) {
	created := false
	svc := newSvc(&mockRepo{
		findContactFn:   func(context.Context, uuid.UUID, string) (*models.Contact, error) { return nil, repository.ErrNotFound },
		createContactFn: func(context.Context, *models.Contact) error { created = true; return nil },
	})
	out, err := svc.EnsureContact(pipeCtx(), uuid.New(), "jid", "Name")
	require.NoError(t, err)
	require.True(t, created)
	require.Equal(t, "jid", out.RemoteJID)
	require.NotEqual(t, uuid.Nil, out.ID)
}

func TestEnsureOpenConversation_ReturnsExisting(t *testing.T) {
	existing := &models.Conversation{ID: uuid.New()}
	created := false
	svc := newSvc(&mockRepo{
		findOpenConvFn: func(context.Context, uuid.UUID) (*models.Conversation, error) { return existing, nil },
		createConvFn:   func(context.Context, *models.Conversation) error { created = true; return nil },
	})
	out, err := svc.EnsureOpenConversation(context.Background(), uuid.New(), uuid.New(), nil)
	require.NoError(t, err)
	require.Equal(t, existing.ID, out.ID)
	require.False(t, created)
}

func TestEnsureOpenConversation_OpensNew(t *testing.T) {
	svc := newSvc(&mockRepo{
		findOpenConvFn: func(context.Context, uuid.UUID) (*models.Conversation, error) { return nil, repository.ErrNotFound },
	})
	out, err := svc.EnsureOpenConversation(pipeCtx(), uuid.New(), uuid.New(), nil)
	require.NoError(t, err)
	require.Equal(t, StateAI, out.State)
	require.NotNil(t, out.LastMessageAt)
}

func TestAppendMessage_Inserts(t *testing.T) {
	conv := &models.Conversation{ID: uuid.New(), CompanyID: uuid.New()}
	bumped := false
	svc := newSvc(&mockRepo{
		updateConvFn: func(context.Context, *models.Conversation) error { bumped = true; return nil },
	})
	msg, inserted, err := svc.AppendMessage(context.Background(), conv, AppendInput{Direction: "in", Content: "hi", ExternalMessageID: "ext-1"})
	require.NoError(t, err)
	require.True(t, inserted)
	require.NotNil(t, msg)
	require.True(t, bumped, "last_message_at must be bumped on insert")
	require.NotNil(t, conv.LastMessageAt)
}

func TestAppendMessage_DuplicateIsNoOp(t *testing.T) {
	conv := &models.Conversation{ID: uuid.New(), CompanyID: uuid.New()}
	bumped := false
	svc := newSvc(&mockRepo{
		insertMessageFn: func(context.Context, *models.Message) error { return repository.ErrDuplicate },
		updateConvFn:    func(context.Context, *models.Conversation) error { bumped = true; return nil },
	})
	msg, inserted, err := svc.AppendMessage(context.Background(), conv, AppendInput{ExternalMessageID: "dup"})
	require.NoError(t, err)
	require.False(t, inserted)
	require.Nil(t, msg)
	require.False(t, bumped, "duplicate must not bump or fan-out")
}

func TestSetState_ClosedSetsClosedAt(t *testing.T) {
	conv := &models.Conversation{ID: uuid.New(), CompanyID: uuid.New(), State: StateAI}
	svc := newSvc(&mockRepo{})
	require.NoError(t, svc.SetState(context.Background(), conv, StateClosed))
	require.Equal(t, StateClosed, conv.State)
	require.NotNil(t, conv.ClosedAt)
}

func TestAssignUser_SetsAssignedUser(t *testing.T) {
	conv := &models.Conversation{ID: uuid.New()}
	uid := uuid.New()
	svc := newSvc(&mockRepo{})
	require.NoError(t, svc.AssignUser(context.Background(), conv, &uid))
	require.Equal(t, &uid, conv.AssignedUserID)
}
