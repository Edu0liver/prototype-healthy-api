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
		s.log.Error("webhook: json unmarshal failed", zap.Error(err))
		return err
	}
	s.log.Debug("webhook: parsed envelope", zap.String("event", env.Event), zap.String("instance", env.Instance))
	if env.Instance == "" {
		s.log.Warn("webhook: empty instance, dropping")
		return nil
	}
	event := normalizeEvent(env.Event)
	s.log.Debug("webhook: normalized event", zap.String("event", event))

	// Route instance -> tenant (system scope; bypasses RLS via SECURITY DEFINER).
	var res repository.Resolved
	var found bool
	if err := s.db.System(ctx, func(ctx context.Context) error {
		var err error
		res, found, err = s.repo.ResolveChannel(ctx, env.Instance)
		return err
	}); err != nil {
		s.log.Error("webhook: ResolveChannel db error", zap.Error(err))
		return err
	}
	if !found {
		s.log.Warn("webhook: unknown instance", zap.String("instance", env.Instance))
		return nil
	}
	s.log.Debug("webhook: instance resolved", zap.String("company_id", res.CompanyID.String()), zap.String("channel_id", res.ChannelID.String()))

	s.audit(ctx, res, event, env)

	switch event {
	case EventMessagesUpsert:
		s.log.Debug("webhook: routing to handleMessage")
		return s.handleMessage(ctx, res, env)
	case EventConnectionUpdate:
		s.log.Debug("webhook: routing to handleConnection")
		return s.handleConnection(ctx, res, env)
	case EventSendMessage, EventMessagesUpdate:
		s.log.Debug("webhook: routing to handleStatus")
		return s.handleStatus(ctx, res, env)
	case EventQRCodeUpdated:
		s.log.Debug("webhook: routing to handleQRCode")
		return s.handleQRCode(ctx, res, env)
	default:
		s.log.Warn("webhook: unhandled event type", zap.String("event", event))
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
		s.log.Error("webhook: handleMessage unmarshal failed", zap.Error(err))
		return err
	}
	s.log.Debug("webhook: handleMessage parsed", zap.String("ext_id", d.Key.ID), zap.String("remote_jid", d.Key.RemoteJID), zap.Bool("from_me", d.Key.FromMe), zap.String("push_name", d.PushName))

	extID := d.Key.ID
	if extID == "" {
		s.log.Warn("webhook: empty message id, dropping")
		return nil
	}

	content := d.text()
	s.log.Debug("webhook: message content", zap.String("content", content), zap.String("message_type", d.MessageType))

	// Idempotency: drop if we've already seen this message id.
	first, err := s.rdb.FirstSeen(ctx, redisx.DedupeKey(extID), 24*time.Hour)
	s.log.Debug("webhook: idempotency check", zap.String("ext_id", extID), zap.Bool("first_seen", first), zap.Error(err))
	if err == nil && !first {
		s.log.Warn("webhook: duplicate message, dropping", zap.String("ext_id", extID))
		return nil
	}

	var job *jobs.InboundJob

	err = s.db.Tenant(ctx, res.CompanyID, func(ctx context.Context) error {
		contact, err := s.conv.EnsureContact(ctx, res.ChannelID, d.Key.RemoteJID, d.PushName)
		if err != nil {
			s.log.Error("webhook: EnsureContact failed", zap.Error(err))
			return err
		}
		s.log.Debug("webhook: contact ensured", zap.String("contact_id", contact.ID.String()))

		conv, err := s.conv.EnsureOpenConversation(ctx, res.ChannelID, contact.ID, nil)
		if err != nil {
			s.log.Error("webhook: EnsureOpenConversation failed", zap.Error(err))
			return err
		}
		s.log.Debug("webhook: conversation ensured", zap.String("conv_id", conv.ID.String()))

		if d.Key.FromMe {
			s.log.Debug("webhook: fromMe=true, applying passive handover block")
			_, _, err := s.conv.AppendMessage(ctx, conv, convsvc.AppendInput{
				Direction: "outbound", SenderType: "human", Content: content,
				ExternalMessageID: extID, Status: "sent",
			})
			if err != nil {
				return err
			}
			// Passive handover (operator replied from the phone): suppress the
			// AI for a TTL window only. We deliberately do NOT flip the
			// conversation to a persistent `human` state — that is reserved for
			// explicit handover (panel take or transfer_to_human). The block is
			// refreshed by each fromMe message, so the AI auto-resumes once the
			// operator stops replying for passiveBlockTTL.
			_ = s.rdb.Block(ctx, conv.ID.String(), passiveBlockTTL)
			return nil
		}

		msg, inserted, err := s.conv.AppendMessage(ctx, conv, convsvc.AppendInput{
			Direction: "inbound", SenderType: "contact", Content: content,
			ExternalMessageID: extID, Status: "received",
		})
		s.log.Debug("webhook: AppendMessage", zap.Bool("inserted", inserted), zap.Error(err))
		if err != nil || !inserted {
			if !inserted {
				s.log.Warn("webhook: message not inserted (duplicate in DB), dropping")
			}
			return err
		}
		s.log.Debug("webhook: message persisted", zap.String("msg_id", msg.ID.String()))
		job = &jobs.InboundJob{
			CompanyID: res.CompanyID.String(), ChannelID: res.ChannelID.String(),
			ConversationID: conv.ID.String(), MessageID: msg.ID.String(), ExternalID: extID,
			Instance: env.Instance, RemoteJID: d.Key.RemoteJID, MessageType: d.MessageType, Content: content,
		}
		return nil
	})
	if err != nil {
		s.log.Error("webhook: tenant transaction failed", zap.Error(err))
		return err
	}

	// Enqueue only after the DB tx committed, so the worker sees the message.
	if job == nil {
		s.log.Warn("webhook: job is nil after transaction, nothing enqueued")
		return nil
	}
	s.log.Debug("webhook: enqueueing job", zap.String("stream", s.cfg.Worker.StreamName), zap.String("conv_id", job.ConversationID))
	if _, err := s.rdb.Enqueue(ctx, s.cfg.Worker.StreamName, job.ToMap()); err != nil {
		s.log.Error("webhook: enqueue failed", zap.Error(err))
		return err
	}
	s.log.Debug("webhook: job enqueued ok")
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
