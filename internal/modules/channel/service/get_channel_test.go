package service

import (
	"context"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGet_OK(t *testing.T) {
	id := uuid.New()
	svc := newSvc(t, &mockRepo{getFn: func(_ context.Context, gotID uuid.UUID) (*models.Channel, error) {
		return &models.Channel{ID: gotID, Name: "WA"}, nil
	}}, &mockEvo{})
	out, err := svc.Get(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "WA", out.Name)
}

func TestGet_NotFoundMapped(t *testing.T) {
	svc := newSvc(t, &mockRepo{getFn: func(context.Context, uuid.UUID) (*models.Channel, error) {
		return nil, repository.ErrNotFound
	}}, &mockEvo{})
	out, err := svc.Get(context.Background(), uuid.New())
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrChannelNotFound)
}
