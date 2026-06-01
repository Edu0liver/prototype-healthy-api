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
	out, err := svc.Connect(context.Background(), uuid.New(), "")
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrNotWhatsApp)
}

func TestConnect_QR(t *testing.T) {
	evo := &mockEvo{connectFn: func(_ context.Context, instance, number string) (*evolution.ConnectResult, error) {
		require.Equal(t, "", number, "no number stored: connect must not pass one")
		return &evolution.ConnectResult{Code: "qr-data"}, nil
	}}
	svc := newSvc(t, &mockRepo{getFn: func(_ context.Context, id uuid.UUID) (*models.Channel, error) {
		return &models.Channel{ID: id, Type: channeladapter.WhatsApp, EvolutionInstanceName: "lumia-x"}, nil
	}}, evo)
	out, err := svc.Connect(context.Background(), uuid.New(), "")
	require.NoError(t, err)
	require.Equal(t, "qr-data", out.QRCode)
	require.Equal(t, StatusConnecting, out.State)
}

func TestConnect_PairingFallsBackToAccountNumber(t *testing.T) {
	evo := &mockEvo{connectFn: func(_ context.Context, instance, number string) (*evolution.ConnectResult, error) {
		require.Equal(t, "+5511888", number, "pairing must fall back to the channel's account id")
		return &evolution.ConnectResult{Code: "qr-data", PairingCode: "PAIR-1"}, nil
	}}
	svc := newSvc(t, &mockRepo{getFn: func(_ context.Context, id uuid.UUID) (*models.Channel, error) {
		return &models.Channel{ID: id, Type: channeladapter.WhatsApp, ExternalAccountID: "+5511888"}, nil
	}}, evo)
	out, err := svc.Connect(context.Background(), uuid.New(), "")
	require.NoError(t, err)
	require.Equal(t, "qr-data", out.QRCode)
	require.Equal(t, "PAIR-1", out.PairingCode)
}

// TestConnect_PairingPollsUntilReady verifies the async pairing code is
// retried: the first connect returns only the QR, a later poll fills pairing.
func TestConnect_PairingPollsUntilReady(t *testing.T) {
	calls := 0
	evo := &mockEvo{connectFn: func(_ context.Context, _, number string) (*evolution.ConnectResult, error) {
		require.Equal(t, "+5511888", number)
		calls++
		if calls == 1 {
			return &evolution.ConnectResult{Code: "qr-data"}, nil // pairing not ready yet
		}
		return &evolution.ConnectResult{PairingCode: "PAIR-2"}, nil
	}}
	svc := newSvc(t, &mockRepo{getFn: func(_ context.Context, id uuid.UUID) (*models.Channel, error) {
		return &models.Channel{ID: id, Type: channeladapter.WhatsApp, ExternalAccountID: "+5511888"}, nil
	}}, evo)
	out, err := svc.Connect(context.Background(), uuid.New(), "")
	require.NoError(t, err)
	require.GreaterOrEqual(t, calls, 2, "must re-poll until pairing code arrives")
	require.Equal(t, "PAIR-2", out.PairingCode)
	require.Equal(t, "qr-data", out.QRCode, "QR from the first call must be preserved")
}

func TestConnect_NotFound(t *testing.T) {
	svc := newSvc(t, &mockRepo{getFn: func(context.Context, uuid.UUID) (*models.Channel, error) {
		return nil, repository.ErrNotFound
	}}, &mockEvo{})
	out, err := svc.Connect(context.Background(), uuid.New(), "")
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrChannelNotFound)
}
