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
