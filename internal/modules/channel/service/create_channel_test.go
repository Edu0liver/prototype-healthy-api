package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/Edu0liver/prototype-healthy-api/pkg/evolution"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createCtx() context.Context {
	return appctx.With(context.Background(), appctx.Identity{CompanyID: uuid.New()})
}

func TestCreate_UnsupportedType(t *testing.T) {
	svc := newSvc(t, &mockRepo{}, &mockEvo{})
	out, err := svc.Create(createCtx(), dto.CreateChannelRequest{Type: "telegram", Name: "x"})
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrUnsupportedType)
}

func TestCreate_Instagram_NoEvolution(t *testing.T) {
	evoCalled := false
	evo := &mockEvo{createInstanceFn: func(context.Context, evolution.CreateInstanceRequest) (*evolution.CreateInstanceResult, error) {
		evoCalled = true
		return &evolution.CreateInstanceResult{}, nil
	}}
	svc := newSvc(t, &mockRepo{}, evo)
	out, err := svc.Create(createCtx(), dto.CreateChannelRequest{Type: channeladapter.Instagram, Name: "IG"})
	require.NoError(t, err)
	require.False(t, evoCalled, "instagram must not provision an evolution instance")
	require.Equal(t, StatusDisconnected, out.Status)
	require.Empty(t, out.EvolutionInstanceName)
}

func TestCreate_WhatsApp_ProvisionsAndEncrypts(t *testing.T) {
	evo := &mockEvo{createInstanceFn: func(_ context.Context, req evolution.CreateInstanceRequest) (*evolution.CreateInstanceResult, error) {
		require.Equal(t, "WHATSAPP-BAILEYS", req.Integration)
		return &evolution.CreateInstanceResult{InstanceID: "inst-123", APIKey: "secret-key"}, nil
	}}
	svc := newSvc(t, &mockRepo{}, evo)

	out, err := svc.Create(createCtx(), dto.CreateChannelRequest{Type: channeladapter.WhatsApp, Name: "WA", Number: "+5511999"})
	require.NoError(t, err)
	require.Equal(t, StatusConnecting, out.Status)
	require.Equal(t, "inst-123", out.EvolutionInstanceID)
	require.Equal(t, "lumia-"+out.ID.String(), out.EvolutionInstanceName)
	require.NotEmpty(t, out.EvolutionAPIKeyEnc)
	require.NotEqual(t, "secret-key", out.EvolutionAPIKeyEnc, "api key must be encrypted at rest")
}

func TestCreate_WhatsApp_EvolutionError(t *testing.T) {
	evo := &mockEvo{createInstanceFn: func(context.Context, evolution.CreateInstanceRequest) (*evolution.CreateInstanceResult, error) {
		return nil, errors.New("evolution unreachable")
	}}
	svc := newSvc(t, &mockRepo{}, evo)
	out, err := svc.Create(createCtx(), dto.CreateChannelRequest{Type: channeladapter.WhatsApp, Name: "WA"})
	require.Nil(t, out)
	require.Error(t, err)
}

func TestCreate_RepoError(t *testing.T) {
	boom := errors.New("insert failed")
	svc := newSvc(t, &mockRepo{createFn: func(context.Context, *models.Channel) error { return boom }}, &mockEvo{})
	out, err := svc.Create(createCtx(), dto.CreateChannelRequest{Type: channeladapter.Instagram, Name: "IG"})
	require.Nil(t, out)
	require.ErrorIs(t, err, boom)
}
