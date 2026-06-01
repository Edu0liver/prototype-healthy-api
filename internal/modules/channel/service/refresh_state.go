package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RefreshState queries Evolution and syncs channels.status.
func (s *Service) RefreshState(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	ch, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}
	if ch.Type != channeladapter.WhatsApp {
		return ch, nil
	}
	state, err := s.evo.ConnectionState(ctx, evoInstance(ch))
	if err != nil {
		// The Evolution instance is gone (e.g. deleted on disconnect). Treat the
		// channel as disconnected instead of erroring, so status polling does not
		// fail on channels that were never/no longer connected.
		s.log.Warn("channel: connection state unavailable, marking disconnected",
			zap.String("channel_id", id.String()), zap.Error(err))
		ch.Status = StatusDisconnected
		if uerr := s.repo.Update(ctx, ch); uerr != nil {
			return nil, uerr
		}
		return ch, nil
	}
	ch.Status = mapState(state)
	if err := s.repo.Update(ctx, ch); err != nil {
		return nil, err
	}
	return ch, nil
}
