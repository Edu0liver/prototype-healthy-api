package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/google/uuid"
)

// EnsureContact finds or creates a contact, refreshing its push name.
func (s *Service) EnsureContact(ctx context.Context, channelID uuid.UUID, remoteJID, pushName string) (*models.Contact, error) {
	c, err := s.repo.FindContact(ctx, channelID, remoteJID)
	if err == nil {
		if pushName != "" && c.PushName != pushName {
			c.PushName = pushName
			_ = s.repo.UpdateContact(ctx, c)
		}
		return c, nil
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}
	c = &models.Contact{
		ID:        uuidV7(),
		CompanyID: appctx.CompanyID(ctx),
		ChannelID: channelID,
		RemoteJID: remoteJID,
		PushName:  pushName,
	}
	if err := s.repo.CreateContact(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}
