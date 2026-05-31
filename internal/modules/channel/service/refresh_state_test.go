package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRefreshState_NotWhatsApp_NoOp(t *testing.T) {
	svc := newSvc(t, &mockRepo{getFn: func(_ context.Context, id uuid.UUID) (*models.Channel, error) {
		return &models.Channel{ID: id, Type: channeladapter.Instagram, Status: StatusConnected}, nil
	}}, &mockEvo{})
	out, err := svc.RefreshState(context.Background(), uuid.New())
	require.NoError(t, err)
	require.Equal(t, StatusConnected, out.Status)
}

func TestRefreshState_MapsEvolutionState(t *testing.T) {
	cases := map[string]string{
		"open":       StatusConnected,
		"connecting": StatusConnecting,
		"close":      StatusDisconnected,
		"weird":      StatusError,
	}
	for evoState, want := range cases {
		t.Run(evoState, func(t *testing.T) {
			evo := &mockEvo{connectionStateFn: func(context.Context, string) (string, error) { return evoState, nil }}
			svc := newSvc(t, &mockRepo{getFn: func(_ context.Context, id uuid.UUID) (*models.Channel, error) {
				return &models.Channel{ID: id, Type: channeladapter.WhatsApp, EvolutionInstanceName: "lumia-x"}, nil
			}}, evo)
			out, err := svc.RefreshState(context.Background(), uuid.New())
			require.NoError(t, err)
			require.Equal(t, want, out.Status)
		})
	}
}

func TestRefreshState_EvolutionError(t *testing.T) {
	evo := &mockEvo{connectionStateFn: func(context.Context, string) (string, error) {
		return "", errors.New("evolution down")
	}}
	svc := newSvc(t, &mockRepo{getFn: func(_ context.Context, id uuid.UUID) (*models.Channel, error) {
		return &models.Channel{ID: id, Type: channeladapter.WhatsApp, EvolutionInstanceName: "lumia-x"}, nil
	}}, evo)
	out, err := svc.RefreshState(context.Background(), uuid.New())
	require.Nil(t, out)
	require.Error(t, err)
}
