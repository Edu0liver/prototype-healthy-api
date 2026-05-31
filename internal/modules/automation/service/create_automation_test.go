package service

import (
	"context"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createCtx() context.Context {
	return appctx.With(context.Background(), appctx.Identity{CompanyID: uuid.New()})
}

func TestCreate_InvalidChannelID(t *testing.T) {
	svc := New(&mockRepo{}, nil)
	out, err := svc.Create(createCtx(), dto.CreateAutomationRequest{ChannelID: "nope", AgentID: uuid.New().String()})
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrChannelNotFound)
}

func TestCreate_InvalidAgentID(t *testing.T) {
	svc := New(&mockRepo{}, nil)
	out, err := svc.Create(createCtx(), dto.CreateAutomationRequest{ChannelID: uuid.New().String(), AgentID: "nope"})
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrAgentNotFound)
}

func TestCreate_ChannelNotInTenant(t *testing.T) {
	svc := New(&mockRepo{channelBelongsFn: func(context.Context, uuid.UUID) (bool, error) { return false, nil }}, nil)
	out, err := svc.Create(createCtx(), dto.CreateAutomationRequest{ChannelID: uuid.New().String(), AgentID: uuid.New().String()})
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrChannelNotFound)
}

func TestCreate_ActiveDefaultsAndReflectsAgent(t *testing.T) {
	chID := uuid.New()
	var reflected *uuid.UUID
	var reflectedChannel uuid.UUID
	repo := &mockRepo{setActiveAgentFn: func(_ context.Context, channelID uuid.UUID, agentID *uuid.UUID) error {
		reflectedChannel, reflected = channelID, agentID
		return nil
	}}
	svc := New(repo, nil)

	agID := uuid.New()
	out, err := svc.Create(createCtx(), dto.CreateAutomationRequest{ChannelID: chID.String(), AgentID: agID.String()})
	require.NoError(t, err)
	require.True(t, out.IsActive, "is_active must default to true")
	require.Equal(t, 8, out.DebounceSeconds, "debounce must default to 8")
	require.Equal(t, chID, reflectedChannel)
	require.NotNil(t, reflected)
	require.Equal(t, agID, *reflected, "active automation must reflect agent onto channel")
}

func TestCreate_InactiveSkipsReflection(t *testing.T) {
	reflected := false
	repo := &mockRepo{setActiveAgentFn: func(context.Context, uuid.UUID, *uuid.UUID) error { reflected = true; return nil }}
	svc := New(repo, nil)

	inactive := false
	out, err := svc.Create(createCtx(), dto.CreateAutomationRequest{
		ChannelID: uuid.New().String(), AgentID: uuid.New().String(), IsActive: &inactive,
	})
	require.NoError(t, err)
	require.False(t, out.IsActive)
	require.False(t, reflected, "inactive automation must not touch channel.active_agent_id")
}

func TestCreate_ActiveExistsMapped(t *testing.T) {
	repo := &mockRepo{createFn: func(context.Context, *models.Automation) error { return repository.ErrActiveExists }}
	svc := New(repo, nil)
	out, err := svc.Create(createCtx(), dto.CreateAutomationRequest{ChannelID: uuid.New().String(), AgentID: uuid.New().String()})
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrActiveExists)
}
