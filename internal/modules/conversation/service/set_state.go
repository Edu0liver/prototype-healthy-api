package service

import (
	"context"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/events"
)

// SetState updates the conversation state (mirror of Redis).
func (s *Service) SetState(ctx context.Context, conv *models.Conversation, state string) error {
	conv.State = state
	if state == StateClosed {
		now := time.Now()
		conv.ClosedAt = &now
	}
	if err := s.repo.UpdateConversation(ctx, conv); err != nil {
		return err
	}
	s.events.Publish(ctx, conv.CompanyID, events.Event{
		Type:           events.TypeState,
		ConversationID: conv.ID.String(),
		Payload:        map[string]any{"state": state},
	})
	return nil
}
