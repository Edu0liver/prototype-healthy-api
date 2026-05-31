package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/google/uuid"
)

// AssignUser assigns an operator to a conversation (handover).
func (s *Service) AssignUser(ctx context.Context, conv *models.Conversation, userID *uuid.UUID) error {
	conv.AssignedUserID = userID
	return s.repo.UpdateConversation(ctx, conv)
}
