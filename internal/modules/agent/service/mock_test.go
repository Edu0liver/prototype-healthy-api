package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/google/uuid"
)

// mockRepo is a function-backed Repository used by the service unit tests.
// Each field, when nil, falls back to a sensible default so tests only wire the
// behaviour they care about.
type mockRepo struct {
	createFn func(ctx context.Context, a *models.Agent) error
	updateFn func(ctx context.Context, a *models.Agent) error
	getFn    func(ctx context.Context, id uuid.UUID) (*models.Agent, error)
	listFn   func(ctx context.Context) ([]models.Agent, error)
	deleteFn func(ctx context.Context, id uuid.UUID) error
}

func (m *mockRepo) Create(ctx context.Context, a *models.Agent) error {
	if m.createFn != nil {
		return m.createFn(ctx, a)
	}
	return nil
}

func (m *mockRepo) Update(ctx context.Context, a *models.Agent) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, a)
	}
	return nil
}

func (m *mockRepo) Get(ctx context.Context, id uuid.UUID) (*models.Agent, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return &models.Agent{ID: id}, nil
}

func (m *mockRepo) List(ctx context.Context) ([]models.Agent, error) {
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
