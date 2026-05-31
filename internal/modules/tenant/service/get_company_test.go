package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetCompany_OK(t *testing.T) {
	id := uuid.New()
	svc := New(&mockRepo{getCompanyFn: func(_ context.Context, gotID uuid.UUID) (*models.Company, error) {
		require.Equal(t, id, gotID)
		return &models.Company{ID: id, Name: "Acme", Slug: "acme"}, nil
	}}, nil)
	out, err := svc.GetCompany(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "Acme", out.Name)
}

func TestGetCompany_NotFoundMapped(t *testing.T) {
	svc := New(&mockRepo{getCompanyFn: func(context.Context, uuid.UUID) (*models.Company, error) {
		return nil, repository.ErrNotFound
	}}, nil)
	out, err := svc.GetCompany(context.Background(), uuid.New())
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrCompanyNotFound)
}

func TestGetCompany_RepoError(t *testing.T) {
	boom := errors.New("db down")
	svc := New(&mockRepo{getCompanyFn: func(context.Context, uuid.UUID) (*models.Company, error) {
		return nil, boom
	}}, nil)
	_, err := svc.GetCompany(context.Background(), uuid.New())
	require.ErrorIs(t, err, boom)
}
