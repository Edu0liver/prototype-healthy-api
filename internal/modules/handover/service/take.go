package service

import (
	"context"

	convsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/google/uuid"
)

// Take transfers a conversation to a human operator (RF-HO-03).
func (s *Service) Take(ctx context.Context, convID, userID uuid.UUID) error {
	conv, err := s.conv.GetConversation(ctx, convID)
	if err != nil {
		return err
	}
	if err := s.conv.SetState(ctx, conv, convsvc.StateHuman); err != nil {
		return err
	}
	if err := s.conv.AssignUser(ctx, conv, &userID); err != nil {
		return err
	}
	_ = s.rdb.SetState(ctx, convID.String(), redisx.StateHuman)
	_ = s.rdb.Block(ctx, convID.String(), blockTTL)
	return nil
}
