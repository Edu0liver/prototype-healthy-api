package service

import (
	"context"

	convsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/google/uuid"
)

// Close ends the conversation (RF-HO-04).
func (s *Service) Close(ctx context.Context, convID uuid.UUID) error {
	conv, err := s.conv.GetConversation(ctx, convID)
	if err != nil {
		return err
	}
	if err := s.conv.SetState(ctx, conv, convsvc.StateClosed); err != nil {
		return err
	}
	_ = s.rdb.SetState(ctx, convID.String(), redisx.StateClosed)
	_ = s.rdb.Unblock(ctx, convID.String())
	return nil
}
