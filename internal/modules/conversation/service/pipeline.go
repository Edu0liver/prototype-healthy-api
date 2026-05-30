package service

import (
	"context"
	"errors"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/events"
	"github.com/google/uuid"
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

// EnsureContact finds or creates a contact, refreshing its push name.
func (s *Service) EnsureContact(ctx context.Context, channelID uuid.UUID, remoteJID, pushName string) (*models.Contact, error) {
	c, err := s.repo.FindContact(ctx, channelID, remoteJID)
	if err == nil {
		if pushName != "" && c.PushName != pushName {
			c.PushName = pushName
			_ = s.repo.UpdateContact(ctx, c)
		}
		return c, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	c = &models.Contact{
		ID:        uuidV7(),
		CompanyID: appctx.CompanyID(ctx),
		ChannelID: channelID,
		RemoteJID: remoteJID,
		PushName:  pushName,
	}
	if err := s.repo.CreateContact(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

// EnsureOpenConversation returns the open conversation for a contact or opens one.
func (s *Service) EnsureOpenConversation(ctx context.Context, channelID, contactID uuid.UUID, agentID *uuid.UUID) (*models.Conversation, error) {
	conv, err := s.repo.FindOpenConversation(ctx, contactID)
	if err == nil {
		return conv, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	now := time.Now()
	conv = &models.Conversation{
		ID:            uuidV7(),
		CompanyID:     appctx.CompanyID(ctx),
		ChannelID:     channelID,
		ContactID:     contactID,
		AgentID:       agentID,
		State:         StateAI,
		LastMessageAt: &now,
		OpenedAt:      now,
	}
	if err := s.repo.CreateConversation(ctx, conv); err != nil {
		return nil, err
	}
	return conv, nil
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

// AssignUser assigns an operator to a conversation (handover).
func (s *Service) AssignUser(ctx context.Context, conv *models.Conversation, userID *uuid.UUID) error {
	conv.AssignedUserID = userID
	return s.repo.UpdateConversation(ctx, conv)
}

// RecentMessages returns the last N messages (chronological) for prompt history.
func (s *Service) RecentMessages(ctx context.Context, convID uuid.UUID, limit int) ([]models.Message, error) {
	return s.repo.RecentMessages(ctx, convID, limit)
}

// MarkStatusByExternalID updates delivery status from Evolution events.
func (s *Service) MarkStatusByExternalID(ctx context.Context, externalID, status string) error {
	return s.repo.UpdateMessageStatusByExternalID(ctx, externalID, status)
}
