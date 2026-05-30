package service

import (
	"context"

	"github.com/google/uuid"
)

// Delete removes an automation and clears the channel's active agent if needed.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	a, err := s.get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if a.IsActive {
		return s.repo.SetChannelActiveAgent(ctx, a.ChannelID, nil)
	}
	return nil
}
