package service

import (
	"context"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGetBranding_OK(t *testing.T) {
	id := uuid.New()
	svc := New(&mockRepo{getBrandingFn: func(_ context.Context, cid uuid.UUID) (*models.CompanyBranding, error) {
		return &models.CompanyBranding{CompanyID: cid, PrimaryColor: "#fff"}, nil
	}}, nil)
	out, err := svc.GetBranding(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, "#fff", out.PrimaryColor)
}

func TestGetBranding_NotFoundMapped(t *testing.T) {
	svc := New(&mockRepo{getBrandingFn: func(context.Context, uuid.UUID) (*models.CompanyBranding, error) {
		return nil, repository.ErrNotFound
	}}, nil)
	out, err := svc.GetBranding(context.Background(), uuid.New())
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrBrandingNotFound)
}
