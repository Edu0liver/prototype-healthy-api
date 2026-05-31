package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/google/uuid"
)

// KBsForAgent returns full knowledge bases linked to an agent.
func (s *Service) KBsForAgent(ctx context.Context, agentID uuid.UUID) ([]models.KnowledgeBase, error) {
	return s.repo.KBsForAgent(ctx, agentID)
}
