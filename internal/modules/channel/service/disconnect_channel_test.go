package service

import (
	"context"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestDisconnect_WhatsApp_TearsDownInstance(t *testing.T) {
	logout, del := false, false
	evo := &mockEvo{
		logoutFn:         func(context.Context, string) error { logout = true; return nil },
		deleteInstanceFn: func(context.Context, string) error { del = true; return nil },
	}
	var saved *models.Channel
	repo := &mockRepo{
		getFn: func(_ context.Context, id uuid.UUID) (*models.Channel, error) {
			return &models.Channel{ID: id, Type: channeladapter.WhatsApp, EvolutionInstanceName: "lumia-x", Status: StatusConnected}, nil
		},
		updateFn: func(_ context.Context, c *models.Channel) error { saved = c; return nil },
	}
	svc := newSvc(t, repo, evo)
	require.NoError(t, svc.Disconnect(context.Background(), uuid.New()))
	require.True(t, logout)
	require.True(t, del)
	require.Equal(t, StatusDisconnected, saved.Status)
}

func TestDisconnect_Instagram_NoEvolution(t *testing.T) {
	logout := false
	evo := &mockEvo{logoutFn: func(context.Context, string) error { logout = true; return nil }}
	var saved *models.Channel
	repo := &mockRepo{
		getFn: func(_ context.Context, id uuid.UUID) (*models.Channel, error) {
			return &models.Channel{ID: id, Type: channeladapter.Instagram, Status: StatusConnected}, nil
		},
		updateFn: func(_ context.Context, c *models.Channel) error { saved = c; return nil },
	}
	svc := newSvc(t, repo, evo)
	require.NoError(t, svc.Disconnect(context.Background(), uuid.New()))
	require.False(t, logout, "instagram disconnect must not call evolution")
	require.Equal(t, StatusDisconnected, saved.Status)
}
