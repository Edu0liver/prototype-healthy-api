// Package services implements Evolution webhook ingestion: validation,
// idempotency, instance->tenant routing, persistence and stream enqueue.
package service

import (
	"context"
	"encoding/json"
	"time"

	convsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/webhook/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/webhook/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/events"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/jobs"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// passiveBlockTTL is how long a fromMe human takeover silences the AI.
const passiveBlockTTL = 30 * time.Minute

// Service processes provider webhooks.
type Service struct {
	db   *database.DB
	rdb  *redisx.Client
	conv *convsvc.Service
	repo *repository.Repository
	pub  *events.Publisher
	cfg  *config.Config
	log  *zap.Logger
}

// New builds the webhook service.
func New(db *database.DB, rdb *redisx.Client, conv *convsvc.Service, repo *repository.Repository, pub *events.Publisher, cfg *config.Config, log *zap.Logger) *Service {
	return &Service{db: db, rdb: rdb, conv: conv, repo: repo, pub: pub, cfg: cfg, log: log}
}

// Process handles one raw Evolution webhook body.
func (s *Service) Process(ctx context.Context, body []byte) error {
	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return err
	}
	if env.Instance == "" {
		return nil
	}
	event := normalizeEvent(env.Event)

	// Route instance -> tenant (system scope; bypasses RLS via SECURITY DEFINER).
	var res repository.Resolved
	var found bool
	if err := s.db.System(ctx, func(ctx context.Context) error {
		var err error
		res, found, err = s.repo.ResolveChannel(ctx, env.Instance)
		return err
	}); err != nil {
		return err
	}
	if !found {
		s.log.Warn("webhook: unknown instance", zap.String("instance", env.Instance))
		return nil
	}

	s.audit(ctx, res, event, env)

	switch event {
	case EventMessagesUpsert:
		return s.handleMessage(ctx, res, env)
	case EventConnectionUpdate:
		return s.handleConnection(ctx, res, env)
	case EventSendMessage, EventMessagesUpdate:
		return s.handleStatus(ctx, res, env)
	case EventQRCodeUpdated:
		return s.handleQRCode(ctx, res, env)
	default:
		return nil
	}
}

func (s *Service) audit(ctx context.Context, res repository.Resolved, event string, env envelope) {
	var payload database.JSONMap
	_ = json.Unmarshal(env.Data, &payload)
	now := time.Now()
	cid, chid := res.CompanyID, res.ChannelID
	_ = s.db.System(ctx, func(ctx context.Context) error {
		return s.repo.InsertEvent(ctx, &models.WebhookEvent{
			ID: uuidV7(), CompanyID: &cid, ChannelID: &chid,
			EventType: event, Payload: payload, ProcessedAt: &now,
		})
	})
}

func (s *Service) handleMessage(ctx context.Context, res repository.Resolved, env envelope) error {
	var d messageData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		return err
	}
	extID := d.Key.ID
	if extID == "" {
		return nil
	}
	// Idempotency: drop if we've already seen this message id.
	first, err := s.rdb.FirstSeen(ctx, redisx.DedupeKey(extID), 24*time.Hour)
	if err == nil && !first {
		return nil
	}

	content := d.text()
	var job *jobs.InboundJob

	err = s.db.Tenant(ctx, res.CompanyID, func(ctx context.Context) error {
		contact, err := s.conv.EnsureContact(ctx, res.ChannelID, d.Key.RemoteJID, d.PushName)
		if err != nil {
			return err
		}
		conv, err := s.conv.EnsureOpenConversation(ctx, res.ChannelID, contact.ID, nil)
		if err != nil {
			return err
		}

		if d.Key.FromMe {
			// Passive handover: a human replied from the phone/WhatsApp Web.
			_, _, err := s.conv.AppendMessage(ctx, conv, convsvc.AppendInput{
				Direction: "outbound", SenderType: "human", Content: content,
				ExternalMessageID: extID, Status: "sent",
			})
			if err != nil {
				return err
			}
			if err := s.conv.SetState(ctx, conv, convsvc.StateHuman); err != nil {
				return err
			}
			_ = s.rdb.SetState(ctx, conv.ID.String(), redisx.StateHuman)
			_ = s.rdb.Block(ctx, conv.ID.String(), passiveBlockTTL)
			return nil
		}

		msg, inserted, err := s.conv.AppendMessage(ctx, conv, convsvc.AppendInput{
			Direction: "inbound", SenderType: "contact", Content: content,
			ExternalMessageID: extID, Status: "received",
		})
		if err != nil || !inserted {
			return err
		}
		job = &jobs.InboundJob{
			CompanyID: res.CompanyID.String(), ChannelID: res.ChannelID.String(),
			ConversationID: conv.ID.String(), MessageID: msg.ID.String(), ExternalID: extID,
			Instance: env.Instance, RemoteJID: d.Key.RemoteJID, MessageType: d.MessageType, Content: content,
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Enqueue only after the DB tx committed, so the worker sees the message.
	if job != nil {
		if _, err := s.rdb.Enqueue(ctx, s.cfg.Worker.StreamName, job.ToMap()); err != nil {
			s.log.Error("webhook: enqueue failed", zap.Error(err))
			return err
		}
	}
	return nil
}

func (s *Service) handleConnection(ctx context.Context, res repository.Resolved, env envelope) error {
	var d connectionData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		return err
	}
	status := mapConnectionState(d.State)
	_ = s.rdb.Set(ctx, "channel:status:"+res.ChannelID.String(), status, 5*time.Minute)
	return s.db.Tenant(ctx, res.CompanyID, func(ctx context.Context) error {
		return s.repo.UpdateChannelStatus(ctx, res.ChannelID, status)
	})
}

func (s *Service) handleStatus(ctx context.Context, res repository.Resolved, env envelope) error {
	var d statusData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		return err
	}
	if d.Key.ID == "" {
		return nil
	}
	status := mapDeliveryStatus(d.Status)
	return s.db.Tenant(ctx, res.CompanyID, func(ctx context.Context) error {
		return s.conv.MarkStatusByExternalID(ctx, d.Key.ID, status)
	})
}

func (s *Service) handleQRCode(ctx context.Context, res repository.Resolved, env envelope) error {
	var d qrCodeData
	if err := json.Unmarshal(env.Data, &d); err != nil {
		return err
	}
	if d.QRCode.Code == "" {
		return nil
	}
	s.pub.Publish(ctx, res.CompanyID, events.Event{
		Type: events.TypeQRUpdate,
		Payload: map[string]any{
			"channel_id": res.ChannelID.String(),
			"qr_code":    d.QRCode.Code,
		},
	})
	return nil
}

func uuidV7() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}
