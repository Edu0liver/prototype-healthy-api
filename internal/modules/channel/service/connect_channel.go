package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/Edu0liver/prototype-healthy-api/pkg/evolution"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// pairingPollAttempts / pairingPollDelay bound how long Connect waits for
// Evolution to produce the (asynchronously generated) pairing code.
const (
	pairingPollAttempts = 4
	pairingPollDelay    = 700 * time.Millisecond
)

// Connect returns the QR code and (when a number is available) the pairing code
// for a WhatsApp channel. Both are returned together so the panel can show them
// at once. If the Evolution instance was deleted (e.g. after Disconnect), it is
// recreated transparently.
func (s *Service) Connect(ctx context.Context, id uuid.UUID, number string) (*dto.ConnectResponse, error) {
	ch, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}
	if ch.Type != channeladapter.WhatsApp {
		return nil, ErrNotWhatsApp
	}

	inst := evoInstance(ch)

	// Disconnect deletes the Evolution instance. Recreate it if it no longer exists.
	state, serr := s.evo.ConnectionState(ctx, inst)
	if serr != nil {
		state = ""
		created, cerr := s.evo.CreateInstance(ctx, evolution.CreateInstanceRequest{
			InstanceName: inst,
			Integration:  "WHATSAPP-BAILEYS",
			QRCode:       true,
			// Recreate in pairing mode so connect returns both pairing + QR codes.
			Number: ch.ExternalAccountID,
			Webhook: &evolution.WebhookConfig{
				URL:     s.cfg.Evolution.WebhookURL,
				Base64:  true,
				Headers: map[string]string{"authorization": "Bearer " + s.cfg.Evolution.WebhookToken},
				Events:  webhookEvents,
			},
		})
		if cerr != nil {
			return nil, fmt.Errorf("channel: recreate instance: %w", cerr)
		}
		enc, cerr := s.cipher.Encrypt(created.APIKey)
		if cerr != nil {
			return nil, cerr
		}
		ch.EvolutionInstanceID = created.InstanceID
		ch.EvolutionAPIKeyEnc = enc
		if cerr := s.repo.Update(ctx, ch); cerr != nil {
			return nil, cerr
		}
	}

	// Already paired: nothing to connect, just sync status.
	if state == "open" {
		ch.Status = StatusConnected
		if err := s.repo.Update(ctx, ch); err != nil {
			return nil, err
		}
		return &dto.ConnectResponse{State: ch.Status}, nil
	}

	// Use the supplied number, falling back to the channel's stored account
	// number. A non-empty number makes Evolution also return a pairing code.
	num := number
	if num == "" {
		num = ch.ExternalAccountID
	}

	res, err := s.evo.Connect(ctx, inst, num)
	if err != nil {
		return nil, fmt.Errorf("channel: connect: %w", err)
	}

	// The pairing code is generated asynchronously by Baileys; the first connect
	// response often has it empty. Poll briefly so the panel gets it in one shot
	// instead of intermittently missing it.
	if num != "" && res.PairingCode == "" {
		res = s.awaitPairing(ctx, inst, num, res)
	}

	// Nothing came back at all: the instance is mid-handshake with an expired/
	// already-delivered code. Restart to force a fresh QRCODE_UPDATED webhook,
	// which reaches the panel over the realtime WebSocket.
	if res.Code == "" && res.PairingCode == "" {
		if rerr := s.evo.Restart(ctx, inst); rerr != nil {
			s.log.Warn("channel: restart to regenerate QR failed", zap.String("instance", inst), zap.Error(rerr))
		}
	}

	ch.Status = StatusConnecting
	if err := s.repo.Update(ctx, ch); err != nil {
		return nil, err
	}
	return &dto.ConnectResponse{QRCode: res.Code, PairingCode: res.PairingCode, State: ch.Status}, nil
}

// awaitPairing re-polls Evolution's connect endpoint until the pairing code is
// available or attempts are exhausted. It preserves any QR code already seen.
func (s *Service) awaitPairing(ctx context.Context, inst, number string, res *evolution.ConnectResult) *evolution.ConnectResult {
	for i := 0; i < pairingPollAttempts && res.PairingCode == ""; i++ {
		select {
		case <-ctx.Done():
			return res
		case <-time.After(pairingPollDelay):
		}
		r, err := s.evo.Connect(ctx, inst, number)
		if err != nil {
			s.log.Warn("channel: pairing poll failed", zap.String("instance", inst), zap.Error(err))
			continue
		}
		if r == nil {
			continue
		}
		if r.Code == "" {
			r.Code = res.Code // keep the QR if the retry dropped it
		}
		res = r
	}
	return res
}
