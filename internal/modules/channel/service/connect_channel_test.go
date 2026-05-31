package service

import (
	"context"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/Edu0liver/prototype-healthy-api/pkg/evolution"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestConnect_NotWhatsApp(t *testing.T) {
	svc := newSvc(t, &mockRepo{getFn: func(_ context.Context, id uuid.UUID) (*models.Channel, error) {
		return &models.Channel{ID: id, Type: channeladapter.Instagram}, nil
	}}, &mockEvo{})
	out, err := svc.Connect(context.Background(), uuid.New(), "qr", "")
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrNotWhatsApp)
}

func TestConnect_QR(t *testing.T) {
	evo := &mockEvo{connectFn: func(_ context.Context, instance, number string) (*evolution.ConnectResult, error) {
		require.Equal(t, "", number, "qr method must not pass a number")
		return &evolution.ConnectResult{Code: "qr-data"}, nil
	}}
	svc := newSvc(t, &mockRepo{getFn: func(_ context.Context, id uuid.UUID) (*models.Channel, error) {
		return &models.Channel{ID: id, Type: channeladapter.WhatsApp, EvolutionInstanceName: "lumia-x"}, nil
	}}, evo)
	out, err := svc.Connect(context.Background(), uuid.New(), "qr", "")
	require.NoError(t, err)
	require.Equal(t, "qr-data", out.QRCode)
	require.Equal(t, StatusConnecting, out.State)
}

func TestConnect_PairingFallsBackToAccountNumber(t *testing.T) {
	evo := &mockEvo{connectFn: func(_ context.Context, instance, number string) (*evolution.ConnectResult, error) {
		require.Equal(t, "+5511888", number, "pairing must fall back to the channel's account id")
		return &evolution.ConnectResult{PairingCode: "PAIR-1"}, nil
	}}
	svc := newSvc(t, &mockRepo{getFn: func(_ context.Context, id uuid.UUID) (*models.Channel, error) {
		return &models.Channel{ID: id, Type: channeladapter.WhatsApp, ExternalAccountID: "+5511888"}, nil
	}}, evo)
	out, err := svc.Connect(context.Background(), uuid.New(), "pairing", "")
	require.NoError(t, err)
	require.Equal(t, "PAIR-1", out.PairingCode)
}

func TestConnect_NotFound(t *testing.T) {
	svc := newSvc(t, &mockRepo{getFn: func(context.Context, uuid.UUID) (*models.Channel, error) {
		return nil, repository.ErrNotFound
	}}, &mockEvo{})
	out, err := svc.Connect(context.Background(), uuid.New(), "qr", "")
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrChannelNotFound)
}
