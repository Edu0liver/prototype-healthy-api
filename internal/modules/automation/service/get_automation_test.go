package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGet_OK(t *testing.T) {
	id := uuid.New()
	svc := New(&mockRepo{getFn: func(_ context.Context, gotID uuid.UUID) (*models.Automation, error) {
		return &models.Automation{ID: gotID, DebounceSeconds: 8}, nil
	}}, nil)
	out, err := svc.Get(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, id, out.ID)
}

func TestGet_NotFoundMapped(t *testing.T) {
	svc := New(&mockRepo{getFn: func(context.Context, uuid.UUID) (*models.Automation, error) {
		return nil, repository.ErrNotFound
	}}, nil)
	out, err := svc.Get(context.Background(), uuid.New())
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrAutomationNotFound)
}

func TestList_OK(t *testing.T) {
	svc := New(&mockRepo{listFn: func(context.Context) ([]models.Automation, error) {
		return []models.Automation{{ID: uuid.New()}}, nil
	}}, nil)
	out, err := svc.List(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
}

func TestList_RepoError(t *testing.T) {
	boom := errors.New("query failed")
	svc := New(&mockRepo{listFn: func(context.Context) ([]models.Automation, error) { return nil, boom }}, nil)
	_, err := svc.List(context.Background())
	require.ErrorIs(t, err, boom)
}
