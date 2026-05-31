package service

import (
	"context"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUpdate_NotFound(t *testing.T) {
	svc := New(&mockRepo{getFn: func(context.Context, uuid.UUID) (*models.Automation, error) {
		return nil, repository.ErrNotFound
	}}, nil)
	out, err := svc.Update(context.Background(), uuid.New(), dto.UpdateAutomationRequest{})
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrAutomationNotFound)
}

func TestUpdate_DeactivateClearsChannelAgent(t *testing.T) {
	chID := uuid.New()
	existing := &models.Automation{ID: uuid.New(), ChannelID: chID, AgentID: uuid.New(), IsActive: true}
	var clearedChannel uuid.UUID
	var clearedAgent *uuid.UUID = ptrUUID(uuid.New()) // sentinel, must become nil
	repo := &mockRepo{
		getFn: func(context.Context, uuid.UUID) (*models.Automation, error) { return existing, nil },
		setActiveAgentFn: func(_ context.Context, channelID uuid.UUID, agentID *uuid.UUID) error {
			clearedChannel, clearedAgent = channelID, agentID
			return nil
		},
	}
	svc := New(repo, nil)

	inactive := false
	out, err := svc.Update(context.Background(), existing.ID, dto.UpdateAutomationRequest{IsActive: &inactive})
	require.NoError(t, err)
	require.False(t, out.IsActive)
	require.Equal(t, chID, clearedChannel)
	require.Nil(t, clearedAgent, "deactivation must clear channel.active_agent_id")
}

func TestUpdate_InvalidAgentRef(t *testing.T) {
	existing := &models.Automation{ID: uuid.New(), IsActive: true}
	repo := &mockRepo{
		getFn:          func(context.Context, uuid.UUID) (*models.Automation, error) { return existing, nil },
		agentBelongsFn: func(context.Context, uuid.UUID) (bool, error) { return false, nil },
	}
	svc := New(repo, nil)
	agID := uuid.New().String()
	out, err := svc.Update(context.Background(), existing.ID, dto.UpdateAutomationRequest{AgentID: &agID})
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrAgentNotFound)
}

func ptrUUID(id uuid.UUID) *uuid.UUID { return &id }
