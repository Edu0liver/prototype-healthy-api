package service

import (
	"context"

	"github.com/google/uuid"
)

// UnlinkAgent removes an agent↔KB link.
func (s *Service) UnlinkAgent(ctx context.Context, agentID, kbID uuid.UUID) error {
	return s.repo.UnlinkAgentKB(ctx, agentID, kbID)
}
