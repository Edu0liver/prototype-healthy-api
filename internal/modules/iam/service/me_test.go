package service

import (
	"context"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newSvc(repo Repository, mailer Mailer) *Service {
	return New(repo, nil, nil, mailer, nil, nil)
}

func TestMe_OK(t *testing.T) {
	id := uuid.New()
	svc := newSvc(&mockRepo{findByIDFn: func(_ context.Context, gotID uuid.UUID) (*models.User, error) {
		require.Equal(t, id, gotID)
		return &models.User{ID: id, Email: "u@x.com"}, nil
	}}, nil)
	out, err := svc.Me(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "u@x.com", out.Email)
}

func TestMe_NotFoundMapped(t *testing.T) {
	svc := newSvc(&mockRepo{findByIDFn: func(context.Context, uuid.UUID) (*models.User, error) {
		return nil, repository.ErrNotFound
	}}, nil)
	out, err := svc.Me(context.Background(), uuid.New())
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrUserNotFound)
}
