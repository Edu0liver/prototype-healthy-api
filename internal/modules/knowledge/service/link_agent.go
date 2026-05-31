package service

import (
	"context"

	"github.com/google/uuid"
)

// LinkAgent links an agent to a knowledge base (N:M, RF-AG-02).
func (s *Service) LinkAgent(ctx context.Context, agentID, kbID uuid.UUID) error {
	if _, err := s.GetKB(ctx, kbID); err != nil {
		return err
	}
	return s.repo.LinkAgentKB(ctx, agentID, kbID)
}
