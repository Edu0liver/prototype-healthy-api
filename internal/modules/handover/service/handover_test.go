package service

import (
	"context"
	"testing"

	convmodels "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	convrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	convsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// newSvc builds a handover service whose conversation dependency is backed by
// convRepo. repo/cipher/adapters are nil — Take/Return/Close never touch them.
func newSvc(convRepo *mockConvRepo) *Service {
	return New(newConvSvc(convRepo), deadRedis(), nil, nil, nil)
}

func TestTake_SetsHumanAndAssigns(t *testing.T) {
	convID, userID := uuid.New(), uuid.New()
	conv := &convmodels.Conversation{ID: convID, CompanyID: uuid.New(), State: convsvc.StateAI}
	svc := newSvc(&mockConvRepo{
		getConvFn: func(context.Context, uuid.UUID) (*convmodels.Conversation, error) { return conv, nil },
	})
	require.NoError(t, svc.Take(context.Background(), convID, userID))
	require.Equal(t, convsvc.StateHuman, conv.State)
	require.NotNil(t, conv.AssignedUserID)
	require.Equal(t, userID, *conv.AssignedUserID)
}

func TestTake_ConversationNotFound(t *testing.T) {
	svc := newSvc(&mockConvRepo{
		getConvFn: func(context.Context, uuid.UUID) (*convmodels.Conversation, error) {
			return nil, convrepo.ErrNotFound
		},
	})
	err := svc.Take(context.Background(), uuid.New(), uuid.New())
	require.ErrorIs(t, err, convsvc.ErrConversationNotFound)
}

func TestReturn_BackToAIAndUnassigns(t *testing.T) {
	conv := &convmodels.Conversation{ID: uuid.New(), CompanyID: uuid.New(), State: convsvc.StateHuman}
	uid := uuid.New()
	conv.AssignedUserID = &uid
	svc := newSvc(&mockConvRepo{
		getConvFn: func(context.Context, uuid.UUID) (*convmodels.Conversation, error) { return conv, nil },
	})
	require.NoError(t, svc.Return(context.Background(), conv.ID))
	require.Equal(t, convsvc.StateAI, conv.State)
	require.Nil(t, conv.AssignedUserID, "returning to AI must clear the operator assignment")
}

func TestClose_SetsClosed(t *testing.T) {
	conv := &convmodels.Conversation{ID: uuid.New(), CompanyID: uuid.New(), State: convsvc.StateHuman}
	svc := newSvc(&mockConvRepo{
		getConvFn: func(context.Context, uuid.UUID) (*convmodels.Conversation, error) { return conv, nil },
	})
	require.NoError(t, svc.Close(context.Background(), conv.ID))
	require.Equal(t, convsvc.StateClosed, conv.State)
	require.NotNil(t, conv.ClosedAt)
}

func TestClose_ConversationNotFound(t *testing.T) {
	svc := newSvc(&mockConvRepo{
		getConvFn: func(context.Context, uuid.UUID) (*convmodels.Conversation, error) {
			return nil, convrepo.ErrNotFound
		},
	})
	require.ErrorIs(t, svc.Close(context.Background(), uuid.New()), convsvc.ErrConversationNotFound)
}

func TestReturn_ConversationNotFound(t *testing.T) {
	svc := newSvc(&mockConvRepo{
		getConvFn: func(context.Context, uuid.UUID) (*convmodels.Conversation, error) {
			return nil, convrepo.ErrNotFound
		},
	})
	require.ErrorIs(t, svc.Return(context.Background(), uuid.New()), convsvc.ErrConversationNotFound)
}

// Reply must refuse when the conversation is not under human control. The guard
// runs before any dispatch dependency (repo/adapters/cipher) is touched.
func TestReply_RejectsWhenNotHuman(t *testing.T) {
	conv := &convmodels.Conversation{ID: uuid.New(), CompanyID: uuid.New(), State: convsvc.StateAI}
	svc := newSvc(&mockConvRepo{
		getConvFn: func(context.Context, uuid.UUID) (*convmodels.Conversation, error) { return conv, nil },
	})
	require.ErrorIs(t, svc.Reply(context.Background(), conv.ID, "hello"), ErrNotHuman)
}

func TestReply_ConversationNotFound(t *testing.T) {
	svc := newSvc(&mockConvRepo{
		getConvFn: func(context.Context, uuid.UUID) (*convmodels.Conversation, error) {
			return nil, convrepo.ErrNotFound
		},
	})
	require.ErrorIs(t, svc.Reply(context.Background(), uuid.New(), "hi"), convsvc.ErrConversationNotFound)
}
