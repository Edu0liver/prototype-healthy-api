package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetConversation_NotFoundMapped(t *testing.T) {
	svc := newSvc(&mockRepo{getConvFn: func(context.Context, uuid.UUID) (*models.Conversation, error) {
		return nil, repository.ErrNotFound
	}})
	out, err := svc.GetConversation(context.Background(), uuid.New())
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrConversationNotFound)
}

func TestList_PassesFilter(t *testing.T) {
	chID := uuid.New()
	svc := newSvc(&mockRepo{listConvFn: func(_ context.Context, f repository.ConversationFilter) ([]models.Conversation, error) {
		require.Equal(t, "human", f.State)
		require.Equal(t, &chID, f.ChannelID)
		return []models.Conversation{{ID: uuid.New()}}, nil
	}})
	out, err := svc.List(context.Background(), repository.ConversationFilter{State: "human", ChannelID: &chID})
	require.NoError(t, err)
	require.Len(t, out, 1)
}

func TestMessages_RepoError(t *testing.T) {
	boom := errors.New("query failed")
	svc := newSvc(&mockRepo{listMessagesFn: func(context.Context, uuid.UUID) ([]models.Message, error) { return nil, boom }})
	_, err := svc.Messages(context.Background(), uuid.New())
	require.ErrorIs(t, err, boom)
}

func TestRecentMessages_PassesLimit(t *testing.T) {
	svc := newSvc(&mockRepo{recentMessagesFn: func(_ context.Context, _ uuid.UUID, limit int) ([]models.Message, error) {
		require.Equal(t, 10, limit)
		return []models.Message{{}}, nil
	}})
	out, err := svc.RecentMessages(context.Background(), uuid.New(), 10)
	require.NoError(t, err)
	require.Len(t, out, 1)
}

func TestMarkStatusByExternalID_OK(t *testing.T) {
	var gotID, gotStatus string
	svc := newSvc(&mockRepo{updateStatusFn: func(_ context.Context, externalID, status string) error {
		gotID, gotStatus = externalID, status
		return nil
	}})
	require.NoError(t, svc.MarkStatusByExternalID(context.Background(), "ext-1", "read"))
	require.Equal(t, "ext-1", gotID)
	require.Equal(t, "read", gotStatus)
}
