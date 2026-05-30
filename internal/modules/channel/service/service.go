// Package service holds the channel use cases (one file per use case):
// provisioning Evolution instances, QR/pairing connection, state sync, disconnect.
package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/pkg/crypto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/evolution"
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

func (s *Service) get(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	ch, err := s.repo.Get(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
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
