package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestList_OK(t *testing.T) {
	want := []models.Agent{{ID: uuid.New()}, {ID: uuid.New()}}
	svc := New(&mockRepo{listFn: func(context.Context) ([]models.Agent, error) { return want, nil }})
	out, err := svc.List(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 2)
}

func TestList_RepoError(t *testing.T) {
	boom := errors.New("query failed")
	svc := New(&mockRepo{listFn: func(context.Context) ([]models.Agent, error) { return nil, boom }})
	out, err := svc.List(context.Background())
	require.Nil(t, out)
	require.ErrorIs(t, err, boom)
}
