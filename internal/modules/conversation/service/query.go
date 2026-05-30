package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	"github.com/google/uuid"
)

// GetConversation loads a conversation by id.
func (s *Service) GetConversation(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	conv, err := s.repo.GetConversation(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrConversationNotFound
	}
	return conv, err
}

// List returns conversations matching the filter.
func (s *Service) List(ctx context.Context, f repository.ConversationFilter) ([]models.Conversation, error) {
	return s.repo.ListConversations(ctx, f)
}

// Messages returns the full message history of a conversation.
func (s *Service) Messages(ctx context.Context, convID uuid.UUID) ([]models.Message, error) {
	return s.repo.ListMessages(ctx, convID)
}
