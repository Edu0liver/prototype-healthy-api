package service

import (
	"context"
	"fmt"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/google/uuid"
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
	state, err := s.evo.ConnectionState(ctx, ch.EvolutionInstanceName)
	if err != nil {
		return nil, fmt.Errorf("channel: connection state: %w", err)
	}
	ch.Status = mapState(state)
	if err := s.repo.Update(ctx, ch); err != nil {
		return nil, err
	}
	return ch, nil
}
