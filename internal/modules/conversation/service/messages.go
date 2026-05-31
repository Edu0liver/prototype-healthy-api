package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/google/uuid"
)

// Messages returns the full message history of a conversation.
func (s *Service) Messages(ctx context.Context, convID uuid.UUID) ([]models.Message, error) {
	return s.repo.ListMessages(ctx, convID)
}
