// Package services holds the channel use cases: provisioning Evolution
// instances, QR/pairing connection, state sync and disconnect.
package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/dtos"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/channeladapter"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/crypto"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/evolution"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Channel status values.
const (
	StatusDisconnected = "disconnected"
	StatusConnecting   = "connecting"
	StatusConnected    = "connected"
	StatusError        = "error"
)

var webhookEvents = []string{
	"QRCODE_UPDATED", "CONNECTION_UPDATE", "MESSAGES_UPSERT", "MESSAGES_UPDATE", "SEND_MESSAGE",
}

// Service implements channel use cases.
type Service struct {
	repo   Repository
	evo    evolution.Client
	cipher *crypto.Cipher
	cfg    *config.Config
	log    *zap.Logger
}

// New builds the channel service.
func New(repo Repository, evo evolution.Client, cipher *crypto.Cipher, cfg *config.Config, log *zap.Logger) *Service {
	return &Service{repo: repo, evo: evo, cipher: cipher, cfg: cfg, log: log}
}

// Create registers a channel and, for WhatsApp, provisions an Evolution instance.
func (s *Service) Create(ctx context.Context, in dtos.CreateChannelRequest) (*models.Channel, error) {
	if in.Type != channeladapter.WhatsApp && in.Type != channeladapter.Instagram {
		return nil, ErrUnsupportedType
	}
	ch := &models.Channel{
		ID:                uuidV7(),
		CompanyID:         appctx.CompanyID(ctx),
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
			Number:       in.Number,
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

// Connect returns QR code or pairing code for a WhatsApp channel.
func (s *Service) Connect(ctx context.Context, id uuid.UUID, method, number string) (*dtos.ConnectResponse, error) {
	ch, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}
	if ch.Type != channeladapter.WhatsApp {
		return nil, ErrNotWhatsApp
	}
	num := ""
	if method == "pairing" {
		num = number
		if num == "" {
			num = ch.ExternalAccountID
		}
	}
	res, err := s.evo.Connect(ctx, ch.EvolutionInstanceName, num)
	if err != nil {
		return nil, fmt.Errorf("channel: connect: %w", err)
	}
	ch.Status = StatusConnecting
	if err := s.repo.Update(ctx, ch); err != nil {
		return nil, err
	}
	return &dtos.ConnectResponse{QRCode: res.Code, PairingCode: res.PairingCode, State: ch.Status}, nil
}

// RefreshState queries Evolution and syncs channels.status.
func (s *Service) RefreshState(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	ch, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}
	if ch.Type != channeladapter.WhatsApp {
		return ch, nil
	}
	state, err := s.evo.ConnectionState(ctx, ch.EvolutionInstanceName)
	if err != nil {
		return nil, fmt.Errorf("channel: connection state: %w", err)
	}
	ch.Status = mapState(state)
	if err := s.repo.Update(ctx, ch); err != nil {
		return nil, err
	}
	return ch, nil
}

// Disconnect logs out and deletes the Evolution instance, marking disconnected.
func (s *Service) Disconnect(ctx context.Context, id uuid.UUID) error {
	ch, err := s.get(ctx, id)
	if err != nil {
		return err
	}
	if ch.Type == channeladapter.WhatsApp && ch.EvolutionInstanceName != "" {
		if err := s.evo.Logout(ctx, ch.EvolutionInstanceName); err != nil {
			s.log.Warn("evolution logout failed", zap.Error(err))
		}
		if err := s.evo.DeleteInstance(ctx, ch.EvolutionInstanceName); err != nil {
			s.log.Warn("evolution delete failed", zap.Error(err))
		}
	}
	ch.Status = StatusDisconnected
	return s.repo.Update(ctx, ch)
}

// Get returns a channel by id.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	return s.get(ctx, id)
}

// List returns all channels in the tenant.
func (s *Service) List(ctx context.Context) ([]models.Channel, error) {
	return s.repo.List(ctx)
}

func (s *Service) get(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	ch, err := s.repo.Get(ctx, id)
	if errors.Is(err, repositories.ErrNotFound) {
		return nil, ErrChannelNotFound
	}
	return ch, err
}

func mapState(evoState string) string {
	switch evoState {
	case "open":
		return StatusConnected
	case "connecting":
		return StatusConnecting
	case "close":
		return StatusDisconnected
	default:
		return StatusError
	}
}

func uuidV7() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}
