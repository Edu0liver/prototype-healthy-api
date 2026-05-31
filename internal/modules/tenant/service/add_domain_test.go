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

func TestAddDomain_OKLowercases(t *testing.T) {
	id := uuid.New()
	var saved *models.CompanyDomain
	svc := New(&mockRepo{addDomainFn: func(_ context.Context, d *models.CompanyDomain) error {
		saved = d
		return nil
	}}, nil)

	out, err := svc.AddDomain(context.Background(), id, dto.AddDomainRequest{Domain: "App.Example.COM", IsPrimary: true})
	require.NoError(t, err)
	require.Equal(t, "app.example.com", out.Domain, "domain must be lowercased")
	require.Equal(t, id, out.CompanyID)
	require.True(t, out.IsPrimary)
	require.NotEqual(t, uuid.Nil, out.ID)
	require.Same(t, saved, out)
}

func TestAddDomain_UniqueViolationMapped(t *testing.T) {
	svc := New(&mockRepo{addDomainFn: func(context.Context, *models.CompanyDomain) error {
		return errors.New("ERROR: duplicate key value violates unique constraint")
	}}, nil)
	out, err := svc.AddDomain(context.Background(), uuid.New(), dto.AddDomainRequest{Domain: "x.com"})
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrDomainTaken)
}

func TestAddDomain_OtherErrorPassthrough(t *testing.T) {
	boom := errors.New("connection refused")
	svc := New(&mockRepo{addDomainFn: func(context.Context, *models.CompanyDomain) error { return boom }}, nil)
	out, err := svc.AddDomain(context.Background(), uuid.New(), dto.AddDomainRequest{Domain: "x.com"})
	require.Nil(t, out)
	require.ErrorIs(t, err, boom)
}
