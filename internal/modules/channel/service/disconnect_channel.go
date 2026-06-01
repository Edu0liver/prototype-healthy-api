package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Disconnect logs out and deletes the Evolution instance, marking disconnected.
func (s *Service) Disconnect(ctx context.Context, id uuid.UUID) error {
	ch, err := s.get(ctx, id)
	if err != nil {
		return err
	}
	if ch.Type == channeladapter.WhatsApp {
		inst := evoInstance(ch)
		if err := s.evo.Logout(ctx, inst); err != nil {
			s.log.Warn("evolution logout failed", zap.Error(err))
		}
		if err := s.evo.DeleteInstance(ctx, inst); err != nil {
			s.log.Warn("evolution delete failed", zap.Error(err))
		}
	}
	ch.Status = StatusDisconnected
	return s.repo.Update(ctx, ch)
}
