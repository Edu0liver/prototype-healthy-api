package service

import "context"

// MarkStatusByExternalID updates delivery status from Evolution events.
func (s *Service) MarkStatusByExternalID(ctx context.Context, externalID, status string) error {
	return s.repo.UpdateMessageStatusByExternalID(ctx, externalID, status)
}
