package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/tenant/infra/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUpdateBranding_OK(t *testing.T) {
	id := uuid.New()
	var saved *models.CompanyBranding
	svc := New(&mockRepo{upsertBrandingFn: func(_ context.Context, b *models.CompanyBranding) error {
		saved = b
		return nil
	}}, nil)

	out, err := svc.UpdateBranding(context.Background(), id, dto.UpdateBrandingRequest{
		LogoURL:      "https://x/logo.png",
		PrimaryColor: "#123456",
	})
	require.NoError(t, err)
	require.Equal(t, id, out.CompanyID)
	require.Equal(t, "https://x/logo.png", out.LogoURL)
	require.Equal(t, "#123456", out.PrimaryColor)
	require.Same(t, saved, out)
}

func TestUpdateBranding_RepoError(t *testing.T) {
	boom := errors.New("upsert failed")
	svc := New(&mockRepo{upsertBrandingFn: func(context.Context, *models.CompanyBranding) error { return boom }}, nil)
	out, err := svc.UpdateBranding(context.Background(), uuid.New(), dto.UpdateBrandingRequest{})
	require.Nil(t, out)
	require.ErrorIs(t, err, boom)
}
