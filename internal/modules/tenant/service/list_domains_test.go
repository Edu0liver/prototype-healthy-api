package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListDomains_OK(t *testing.T) {
	id := uuid.New()
	svc := New(&mockRepo{listDomainsFn: func(_ context.Context, cid uuid.UUID) ([]models.CompanyDomain, error) {
		require.Equal(t, id, cid)
		return []models.CompanyDomain{{Domain: "a.com"}, {Domain: "b.com"}}, nil
	}}, nil)
	out, err := svc.ListDomains(context.Background(), id)
	require.NoError(t, err)
	require.Len(t, out, 2)
}

func TestListDomains_RepoError(t *testing.T) {
	boom := errors.New("query failed")
	svc := New(&mockRepo{listDomainsFn: func(context.Context, uuid.UUID) ([]models.CompanyDomain, error) {
		return nil, boom
	}}, nil)
	_, err := svc.ListDomains(context.Background(), uuid.New())
	require.ErrorIs(t, err, boom)
}
