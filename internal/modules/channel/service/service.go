// Package service holds the channel use cases (one file per use case):
// provisioning Evolution instances, QR/pairing connection, state sync, disconnect.
package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
	bill   QuotaGuard
	log    *zap.Logger
}

// New builds the channel service. Billing enforcement defaults to a no-op; the
// fx module wires the real guard via WithBilling so unit tests need no billing.
func New(repo Repository, evo evolution.Client, cipher *crypto.Cipher, cfg *config.Config, log *zap.Logger) *Service {
	return &Service{repo: repo, evo: evo, cipher: cipher, cfg: cfg, bill: noopQuota{}, log: log}
}

// WithBilling installs the billing quota guard (production wiring).
func (s *Service) WithBilling(b QuotaGuard) *Service { s.bill = b; return s }

func (s *Service) get(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	ch, err := s.repo.Get(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrChannelNotFound
	}
	return ch, err
}

// evoInstance returns the Evolution instance name for a channel.
// If the stored name is empty (legacy channels) or missing the prefix,
// it is derived from the channel ID so all Evolution calls remain consistent.
func evoInstance(ch *models.Channel) string {
	const prefix = "lumia-"
	name := ch.EvolutionInstanceName
	if name == "" {
		return fmt.Sprintf("%s%s", prefix, ch.ID.String())
	}
	if !strings.HasPrefix(name, prefix) {
		return prefix + name
	}
	return name
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
