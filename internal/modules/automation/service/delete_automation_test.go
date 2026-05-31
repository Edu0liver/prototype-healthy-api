package service

import (
	"context"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDelete_ActiveClearsChannelAgent(t *testing.T) {
	chID := uuid.New()
	existing := &models.Automation{ID: uuid.New(), ChannelID: chID, IsActive: true}
	deleted := false
	cleared := false
	repo := &mockRepo{
		getFn:    func(context.Context, uuid.UUID) (*models.Automation, error) { return existing, nil },
		deleteFn: func(context.Context, uuid.UUID) error { deleted = true; return nil },
		setActiveAgentFn: func(_ context.Context, channelID uuid.UUID, agentID *uuid.UUID) error {
			cleared = true
			require.Equal(t, chID, channelID)
			require.Nil(t, agentID)
			return nil
		},
	}
	svc := New(repo, nil)
	require.NoError(t, svc.Delete(context.Background(), existing.ID))
	require.True(t, deleted)
	require.True(t, cleared, "deleting an active automation must clear channel.active_agent_id")
}

func TestDelete_InactiveSkipsClear(t *testing.T) {
	existing := &models.Automation{ID: uuid.New(), IsActive: false}
	cleared := false
	repo := &mockRepo{
		getFn:            func(context.Context, uuid.UUID) (*models.Automation, error) { return existing, nil },
		setActiveAgentFn: func(context.Context, uuid.UUID, *uuid.UUID) error { cleared = true; return nil },
	}
	svc := New(repo, nil)
	require.NoError(t, svc.Delete(context.Background(), existing.ID))
	require.False(t, cleared)
}

func TestDelete_NotFound(t *testing.T) {
	svc := New(&mockRepo{getFn: func(context.Context, uuid.UUID) (*models.Automation, error) {
		return nil, repository.ErrNotFound
	}}, nil)
	require.ErrorIs(t, svc.Delete(context.Background(), uuid.New()), ErrAutomationNotFound)
}
