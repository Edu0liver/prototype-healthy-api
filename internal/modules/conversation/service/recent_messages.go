package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/google/uuid"
)

// RecentMessages returns the last N messages (chronological) for prompt history.
func (s *Service) RecentMessages(ctx context.Context, convID uuid.UUID, limit int) ([]models.Message, error) {
	return s.repo.RecentMessages(ctx, convID, limit)
}
