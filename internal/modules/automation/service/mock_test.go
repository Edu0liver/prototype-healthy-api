package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/google/uuid"
)

// mockRepo is a function-backed automation Repository. Unset funcs default to
// the permissive happy path (refs belong to tenant, writes succeed).
type mockRepo struct {
	createFn         func(ctx context.Context, a *models.Automation) error
	updateFn         func(ctx context.Context, a *models.Automation) error
	getFn            func(ctx context.Context, id uuid.UUID) (*models.Automation, error)
	listFn           func(ctx context.Context) ([]models.Automation, error)
	deleteFn         func(ctx context.Context, id uuid.UUID) error
	channelBelongsFn func(ctx context.Context, channelID uuid.UUID) (bool, error)
	agentBelongsFn   func(ctx context.Context, agentID uuid.UUID) (bool, error)
	setActiveAgentFn func(ctx context.Context, channelID uuid.UUID, agentID *uuid.UUID) error
}

func (m *mockRepo) Create(ctx context.Context, a *models.Automation) error {
	if m.createFn != nil {
		return m.createFn(ctx, a)
	}
	return nil
}

func (m *mockRepo) Update(ctx context.Context, a *models.Automation) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, a)
	}
	return nil
}

func (m *mockRepo) Get(ctx context.Context, id uuid.UUID) (*models.Automation, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return &models.Automation{ID: id}, nil
}

func (m *mockRepo) List(ctx context.Context) ([]models.Automation, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

func (m *mockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockRepo) ChannelBelongsToTenant(ctx context.Context, channelID uuid.UUID) (bool, error) {
	if m.channelBelongsFn != nil {
		return m.channelBelongsFn(ctx, channelID)
	}
	return true, nil
}

func (m *mockRepo) AgentBelongsToTenant(ctx context.Context, agentID uuid.UUID) (bool, error) {
	if m.agentBelongsFn != nil {
		return m.agentBelongsFn(ctx, agentID)
	}
	return true, nil
}

func (m *mockRepo) SetChannelActiveAgent(ctx context.Context, channelID uuid.UUID, agentID *uuid.UUID) error {
	if m.setActiveAgentFn != nil {
		return m.setActiveAgentFn(ctx, channelID, agentID)
	}
	return nil
}
