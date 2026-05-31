package service

import (
	"context"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGet_OK(t *testing.T) {
	id := uuid.New()
	svc := New(&mockRepo{getFn: func(_ context.Context, gotID uuid.UUID) (*models.Agent, error) {
		require.Equal(t, id, gotID)
		return &models.Agent{ID: id, Name: "Bot"}, nil
	}})
	out, err := svc.Get(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "Bot", out.Name)
}

func TestGet_NotFoundMapped(t *testing.T) {
	svc := New(&mockRepo{getFn: func(context.Context, uuid.UUID) (*models.Agent, error) {
		return nil, repository.ErrNotFound
	}})
	out, err := svc.Get(context.Background(), uuid.New())
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrAgentNotFound, "repo ErrNotFound must map to domain ErrAgentNotFound")
}
