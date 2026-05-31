package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/google/uuid"
)

// LinkAgent links an agent to a knowledge base (N:M, RF-AG-02).
func (s *Service) LinkAgent(ctx context.Context, agentID, kbID uuid.UUID) error {
	if _, err := s.GetKB(ctx, kbID); err != nil {
		return err
	}
	return s.repo.LinkAgentKB(ctx, agentID, kbID)
}

// UnlinkAgent removes an agent↔KB link.
func (s *Service) UnlinkAgent(ctx context.Context, agentID, kbID uuid.UUID) error {
	return s.repo.UnlinkAgentKB(ctx, agentID, kbID)
}

// KBsForAgent returns full knowledge bases linked to an agent.
func (s *Service) KBsForAgent(ctx context.Context, agentID uuid.UUID) ([]models.KnowledgeBase, error) {
	return s.repo.KBsForAgent(ctx, agentID)
}
