package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListUsers_OK(t *testing.T) {
	svc := newSvc(&mockRepo{listByCompanyFn: func(context.Context) ([]models.User, error) {
		return []models.User{{ID: uuid.New()}, {ID: uuid.New()}}, nil
	}}, nil)
	out, err := svc.ListUsers(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 2)
}

func TestListUsers_RepoError(t *testing.T) {
	boom := errors.New("query failed")
	svc := newSvc(&mockRepo{listByCompanyFn: func(context.Context) ([]models.User, error) {
		return nil, boom
	}}, nil)
	_, err := svc.ListUsers(context.Background())
	require.ErrorIs(t, err, boom)
}
