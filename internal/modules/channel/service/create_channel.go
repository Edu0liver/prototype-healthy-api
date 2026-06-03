package service

import (
	"context"
	"fmt"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/Edu0liver/prototype-healthy-api/pkg/evolution"
)

// Create registers a channel and, for WhatsApp, provisions an Evolution instance.
func (s *Service) Create(ctx context.Context, in dto.CreateChannelRequest) (*models.Channel, error) {
	if in.Type != channeladapter.WhatsApp && in.Type != channeladapter.Instagram {
		return nil, ErrUnsupportedType
	}
	companyID := appctx.CompanyID(ctx)
	if err := s.bill.EnsureResource(ctx, companyID, "channels"); err != nil {
		return nil, err
	}
	ch := &models.Channel{
		ID:                uuidV7(),
		CompanyID:         companyID,
		Type:              in.Type,
		Name:              in.Name,
		ExternalAccountID: in.Number,
		Status:            StatusDisconnected,
		Metadata:          map[string]any{},
	}

	if in.Type == channeladapter.WhatsApp {
		instanceName := fmt.Sprintf("lumia-%s", ch.ID.String())
		res, err := s.evo.CreateInstance(ctx, evolution.CreateInstanceRequest{
			InstanceName: instanceName,
			Integration:  "WHATSAPP-BAILEYS",
			QRCode:       true,
			// Passing the number puts the instance in pairing mode, which (on
			// Evolution v2.2.3) makes connect return BOTH the pairing code and
			// the QR code, so the panel can offer either. Omitting it yields a
			// QR-only instance (pairingCode always null).
			Number: in.Number,
			Webhook: &evolution.WebhookConfig{
				URL:     s.cfg.Evolution.WebhookURL,
				Base64:  true,
				Headers: map[string]string{"authorization": "Bearer " + s.cfg.Evolution.WebhookToken},
				Events:  webhookEvents,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("channel: provision instance: %w", err)
		}
		enc, err := s.cipher.Encrypt(res.APIKey)
		if err != nil {
			return nil, err
		}
		ch.EvolutionInstanceName = instanceName
		ch.EvolutionInstanceID = res.InstanceID
		ch.EvolutionAPIKeyEnc = enc
		ch.Status = StatusConnecting
	}

	if err := s.repo.Create(ctx, ch); err != nil {
		return nil, err
	}
	return ch, nil
}
