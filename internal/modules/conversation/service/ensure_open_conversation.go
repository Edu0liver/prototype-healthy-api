package service

import (
	"context"
	"errors"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/google/uuid"
)

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
