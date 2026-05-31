package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDelete_OK(t *testing.T) {
	id := uuid.New()
	var gotID uuid.UUID
	svc := New(&mockRepo{deleteFn: func(_ context.Context, did uuid.UUID) error {
		gotID = did
		return nil
	}})
	require.NoError(t, svc.Delete(context.Background(), id))
	require.Equal(t, id, gotID)
}

func TestDelete_NotFoundMapped(t *testing.T) {
	svc := New(&mockRepo{deleteFn: func(context.Context, uuid.UUID) error { return repository.ErrNotFound }})
	err := svc.Delete(context.Background(), uuid.New())
	require.ErrorIs(t, err, ErrAgentNotFound)
}

func TestDelete_RepoError(t *testing.T) {
	boom := errors.New("delete failed")
	svc := New(&mockRepo{deleteFn: func(context.Context, uuid.UUID) error { return boom }})
	err := svc.Delete(context.Background(), uuid.New())
	require.ErrorIs(t, err, boom)
}
