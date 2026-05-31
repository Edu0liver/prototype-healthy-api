package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
)

// List returns conversations matching the filter.
func (s *Service) List(ctx context.Context, f repository.ConversationFilter) ([]models.Conversation, error) {
	return s.repo.ListConversations(ctx, f)
}
