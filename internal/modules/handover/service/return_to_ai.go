package service

import (
	"context"

	convsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/google/uuid"
)

// Return hands the conversation back to the AI (RF-HO-04).
func (s *Service) Return(ctx context.Context, convID uuid.UUID) error {
	conv, err := s.conv.GetConversation(ctx, convID)
	if err != nil {
		return err
	}
	if err := s.conv.SetState(ctx, conv, convsvc.StateAI); err != nil {
		return err
	}
	_ = s.conv.AssignUser(ctx, conv, nil)
	_ = s.rdb.SetState(ctx, convID.String(), redisx.StateAI)
	_ = s.rdb.Unblock(ctx, convID.String())
	return nil
}
