package service

import (
	"context"
	"errors"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/events"
)

// AppendInput carries the fields for AppendMessage.
type AppendInput struct {
	Direction         string
	SenderType        string
	Content           string
	Media             map[string]any
	ExternalMessageID string
	Status            string
}

// AppendMessage persists a message idempotently and bumps last_message_at.
// Returns inserted=false when the external id was already seen.
func (s *Service) AppendMessage(ctx context.Context, conv *models.Conversation, in AppendInput) (*models.Message, bool, error) {
	m := &models.Message{
		ID:                uuidV7(),
		CompanyID:         conv.CompanyID,
		ConversationID:    conv.ID,
		Direction:         in.Direction,
		SenderType:        in.SenderType,
		Content:           in.Content,
		Media:             in.Media,
		ExternalMessageID: in.ExternalMessageID,
		Status:            in.Status,
	}
	err := s.repo.InsertMessage(ctx, m)
	if errors.Is(err, repository.ErrDuplicate) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	now := time.Now()
	conv.LastMessageAt = &now
	if err := s.repo.UpdateConversation(ctx, conv); err != nil {
		return nil, true, err
	}
	s.events.Publish(ctx, conv.CompanyID, events.Event{
		Type:           events.TypeMessage,
		ConversationID: conv.ID.String(),
		Payload: map[string]any{
			"id": m.ID.String(), "direction": m.Direction, "sender_type": m.SenderType,
			"content": m.Content, "status": m.Status, "created_at": m.CreatedAt,
		},
	})
	return m, true, nil
}
