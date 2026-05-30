package service

import (
	"context"
	"fmt"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/google/uuid"
)

// Connect returns QR code or pairing code for a WhatsApp channel.
func (s *Service) Connect(ctx context.Context, id uuid.UUID, method, number string) (*dto.ConnectResponse, error) {
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
	return &dto.ConnectResponse{QRCode: res.Code, PairingCode: res.PairingCode, State: ch.Status}, nil
}
