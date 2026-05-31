package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUpdate_PartialFields(t *testing.T) {
	id := uuid.New()
	existing := &models.Agent{ID: id, Name: "Old", SystemPrompt: "old", Model: "gpt-4o-mini", Status: "draft"}
	var savedName, savedStatus string
	svc := New(&mockRepo{
		getFn: func(_ context.Context, _ uuid.UUID) (*models.Agent, error) { return existing, nil },
		updateFn: func(_ context.Context, a *models.Agent) error {
			savedName, savedStatus = a.Name, a.Status
			return nil
		},
	})

	newName := "New"
	newStatus := "active"
	out, err := svc.Update(context.Background(), id, dto.UpdateAgentRequest{Name: &newName, Status: &newStatus})
	require.NoError(t, err)
	require.Equal(t, "New", out.Name)
	require.Equal(t, "active", out.Status)
	// Untouched field stays the same.
	require.Equal(t, "old", out.SystemPrompt)
	require.Equal(t, "New", savedName)
	require.Equal(t, "active", savedStatus)
}

func TestUpdate_NotFound(t *testing.T) {
	svc := New(&mockRepo{getFn: func(context.Context, uuid.UUID) (*models.Agent, error) {
		return nil, repository.ErrNotFound
	}})
	out, err := svc.Update(context.Background(), uuid.New(), dto.UpdateAgentRequest{})
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrAgentNotFound)
}

func TestUpdate_RepoSaveError(t *testing.T) {
	boom := errors.New("save failed")
	svc := New(&mockRepo{
		getFn:    func(context.Context, uuid.UUID) (*models.Agent, error) { return &models.Agent{}, nil },
		updateFn: func(context.Context, *models.Agent) error { return boom },
	})
	out, err := svc.Update(context.Background(), uuid.New(), dto.UpdateAgentRequest{})
	require.Nil(t, out)
	require.ErrorIs(t, err, boom)
}
