// Package services implements the orchestration pipeline: the core inbound
// processing described in PRD §2.6 plus the PROMPT.md extras (audio
// transcription, output humanization, handover triggers).
package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	convmodels "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	convsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	knowsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/orchestration/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/jobs"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/crypto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/evolution"
	"github.com/Edu0liver/prototype-healthy-api/pkg/openai"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	bufferTTL      = 60 * time.Second
	lockTTL        = 90 * time.Second
	ragTopK        = 10
	historyLimit   = 15
	minSendDelayMs = 2000
	maxSendDelayMs = 3000
)

// transferToHumanTool is the function-calling tool that triggers handover.
var transferToHumanTool = openai.Tool{
	Type: "function",
	Function: openai.FunctionDef{
		Name:        "transfer_to_human",
		Description: "Transfer the conversation to a human operator when the user explicitly asks for a human, is frustrated, or the request is beyond the assistant's scope.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"reason":{"type":"string"}}}`),
	},
}

// Service runs the orchestration pipeline.
type Service struct {
	db       *database.DB
	rdb      *redisx.Client
	cfg      *config.Config
	log      *zap.Logger
	repo     *repository.Repository
	conv     *convsvc.Service
	know     *knowsvc.Service
	evo      evolution.Client
	oa       openai.Client
	cipher   *crypto.Cipher
	adapters *channeladapter.Registry
}

// New builds the orchestration service.
func New(db *database.DB, rdb *redisx.Client, cfg *config.Config, log *zap.Logger,
	repo *repository.Repository, conv *convsvc.Service, know *knowsvc.Service,
	evo evolution.Client, oa openai.Client, cipher *crypto.Cipher, adapters *channeladapter.Registry) *Service {
	return &Service{db: db, rdb: rdb, cfg: cfg, log: log, repo: repo, conv: conv,
		know: know, evo: evo, oa: oa, cipher: cipher, adapters: adapters}
}

// Process handles a single inbound job end-to-end.
func (s *Service) Process(ctx context.Context, job jobs.InboundJob) error {
	s.log.Debug("pipeline: job received", zap.String("conv_id", job.ConversationID), zap.String("channel_id", job.ChannelID), zap.String("content", job.Content))
	companyID, err := uuid.Parse(job.CompanyID)
	if err != nil {
		return err
	}
	channelID, _ := uuid.Parse(job.ChannelID)
	convID, _ := uuid.Parse(job.ConversationID)

	// Load channel creds + active agent (tenant scope).
	var creds *repository.ChannelCreds
	var agentCfg *repository.AgentConfig
	if err := s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
		creds, err = s.repo.LoadChannelCreds(ctx, channelID)
		if err != nil {
			s.log.Error("pipeline: LoadChannelCreds failed", zap.Error(err))
			return err
		}
		s.log.Debug("pipeline: channel creds loaded", zap.String("instance", creds.InstanceName), zap.String("type", creds.Type))
		agentCfg, err = s.repo.LoadActiveAgent(ctx, channelID)
		if errors.Is(err, repository.ErrNoActiveAgent) {
			s.log.Warn("pipeline: no active agent for channel, dropping", zap.String("channel_id", channelID.String()))
			return nil // no agent: nothing to do
		}
		return err
	}); err != nil {
		return err
	}
	if agentCfg == nil {
		s.log.Warn("pipeline: agentCfg is nil, dropping job")
		return nil // no active agent on this channel — drop
	}
	s.log.Debug("pipeline: agent config loaded", zap.String("agent_id", agentCfg.AgentID.String()), zap.String("model", agentCfg.Model), zap.Int("debounce_s", agentCfg.DebounceSeconds))
	apiKey, _ := s.cipher.Decrypt(creds.APIKeyEnc)

	// Audio transcription (PROMPT 5): fetch base64 → Whisper → use transcript.
	content := job.Content
	if job.MessageType == "audioMessage" {
		if t := s.transcribe(ctx, creds.InstanceName, apiKey, job.ExternalID); t != "" {
			content = t
		}
	}

	// Buffer the fragment (debounce). Always push so nothing is lost.
	_ = s.rdb.PushBuffer(ctx, convID.String(), content, bufferTTL)
	s.log.Debug("pipeline: fragment buffered", zap.String("conv_id", convID.String()))

	// Serialize per-conversation processing (Redlock).
	lock, err := s.rdb.AcquireLock(ctx, redisx.LockKey(convID.String()), lockTTL)
	if errors.Is(err, redisx.ErrLockNotAcquired) {
		s.log.Debug("pipeline: lock not acquired, another worker owns conv")
		return nil // another worker owns this conversation; our fragment is buffered
	}
	if err != nil {
		s.log.Error("pipeline: AcquireLock error", zap.Error(err))
		return err
	}
	s.log.Debug("pipeline: lock acquired")
	defer lock.Release(ctx)

	// Debounce window: wait for more fragments, then aggregate.
	debounce := time.Duration(orInt(agentCfg.DebounceSeconds, s.cfg.Worker.DebounceSeconds)) * time.Second
	s.log.Debug("pipeline: debounce wait", zap.Duration("duration", debounce))
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(debounce):
	}
	fragments, err := s.rdb.DrainBuffer(ctx, convID.String())
	if err != nil {
		s.log.Error("pipeline: DrainBuffer error", zap.Error(err))
		return err
	}
	s.log.Debug("pipeline: buffer drained", zap.Int("fragments", len(fragments)))
	if len(fragments) == 0 {
		s.log.Warn("pipeline: no fragments after drain, already processed by peer")
		return nil // already processed by a peer
	}
	aggregated := strings.TrimSpace(strings.Join(fragments, "\n"))
	if aggregated == "" {
		s.log.Warn("pipeline: aggregated content empty, dropping")
		return nil
	}
	s.log.Debug("pipeline: aggregated content", zap.String("content", aggregated))

	// Load conversation + resolve effective state (Redis, PG fallback).
	conv, state, err := s.loadConversation(ctx, companyID, convID)
	if err != nil {
		s.log.Error("pipeline: loadConversation error", zap.Error(err))
		return err
	}
	s.log.Debug("pipeline: conversation state", zap.String("state", state))
	if state == redisx.StateHuman {
		s.log.Warn("pipeline: state=human, AI staying silent")
		return nil // human is in control — IA stays silent (RF-HO-02)
	}
	if state == redisx.StateClosed {
		// Reopen policy (PRD §2.6b): a closed conversation that receives a new
		// inbound is reopened under AI control. New threads are normally created
		// at the webhook layer (FindOpenConversation excludes closed); this
		// guards a stale `closed` mirror in Redis so we don't answer on a thread
		// the panel still shows as closed.
		s.log.Debug("pipeline: state=closed, reopening under AI control")
		state = redisx.StateAI
		_ = s.rdb.SetState(ctx, convID.String(), redisx.StateAI)
	}
	blocked, _ := s.rdb.IsBlocked(ctx, convID.String())
	s.log.Debug("pipeline: block check", zap.Bool("blocked", blocked))
	if blocked {
		s.log.Warn("pipeline: conversation blocked, dropping")
		return nil
	}

	// Business hours (automations.business_hours): outside the configured window
	// the AI does not engage. Send the fallback message if set, mark the inbound
	// as read, and stop — without calling the LLM or changing handover state.
	if !withinBusinessHours(agentCfg.BusinessHours, time.Now()) {
		s.log.Debug("pipeline: outside business hours, not engaging")
		s.sendFallback(ctx, creds, apiKey, job, agentCfg.FallbackMessage)
		s.markRead(ctx, creds, apiKey, job)
		return nil
	}

	// Keyword-based handover (RF-HO-01).
	if agentCfg.HandoverEnabled && containsKeyword(aggregated, agentCfg.HandoverKeywords) {
		s.log.Debug("pipeline: handover keyword matched")
		return s.handover(ctx, companyID, conv)
	}

	// RAG retrieval (tenant scope) + recent history.
	var ragChunks []string
	var history []convmodels.Message
	if err := s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
		hits, herr := s.know.Retrieve(ctx, agentCfg.AgentID, aggregated, ragTopK)
		if herr != nil {
			s.log.Warn("rag retrieve failed", zap.Error(herr))
		}
		s.log.Debug("pipeline: RAG hits", zap.Int("count", len(hits)))
		for _, h := range hits {
			ragChunks = append(ragChunks, h.Content)
		}
		history, _ = s.conv.RecentMessages(ctx, convID, historyLimit)
		s.log.Debug("pipeline: history loaded", zap.Int("messages", len(history)))
		return nil
	}); err != nil {
		return err
	}

	// OpenAI chat with function calling.
	s.log.Debug("pipeline: calling OpenAI", zap.String("model", agentCfg.Model), zap.Int("rag_chunks", len(ragChunks)))
	result, err := s.oa.Chat(ctx, openai.ChatRequest{
		Model:       agentCfg.Model,
		Messages:    buildMessages(agentCfg.SystemPrompt, ragChunks, history, aggregated),
		Tools:       toolset(agentCfg.HandoverEnabled),
		Temperature: float32(agentCfg.Temperature),
		MaxTokens:   agentCfg.MaxOutputTokens,
	})
	if err != nil {
		s.log.Error("openai chat failed", zap.Error(err))
		s.sendFallback(ctx, creds, apiKey, job, agentCfg.FallbackMessage)
		return nil
	}
	s.log.Debug("pipeline: OpenAI response", zap.String("content", result.Content), zap.Int("tool_calls", len(result.ToolCalls)))
	if hasTransfer(result.ToolCalls) {
		s.log.Debug("pipeline: transfer_to_human tool called")
		return s.handover(ctx, companyID, conv)
	}

	// Humanize + send.
	parts := humanize(result.Content)
	s.log.Debug("pipeline: humanized parts", zap.Int("count", len(parts)))
	if len(parts) == 0 {
		s.log.Warn("pipeline: no parts after humanize, nothing to send")
		return nil
	}
	s.log.Debug("pipeline: calling deliver")
	s.deliver(ctx, companyID, creds, apiKey, job, conv, parts)

	// Read receipt + mirror state.
	s.markRead(ctx, creds, apiKey, job)
	_ = s.rdb.SetState(ctx, convID.String(), redisx.StateAI)
	s.log.Debug("pipeline: done")
	return nil
}

func (s *Service) transcribe(ctx context.Context, instance, apiKey, externalID string) string {
	b64, _, err := s.evo.GetMediaBase64(ctx, instance, apiKey, externalID)
	if err != nil {
		s.log.Warn("get media base64 failed", zap.Error(err))
		return ""
	}
	audio, err := base64.StdEncoding.DecodeString(strings.TrimSpace(b64))
	if err != nil {
		return ""
	}
	txt, err := s.oa.Transcribe(ctx, audio, "audio.ogg")
	if err != nil {
		s.log.Warn("whisper transcription failed", zap.Error(err))
		return ""
	}
	return txt
}

// loadConversation loads the conversation and resolves its state, repopulating
// Redis from Postgres on a cache miss (RF-HO-05).
func (s *Service) loadConversation(ctx context.Context, companyID, convID uuid.UUID) (*convmodels.Conversation, string, error) {
	var conv *convmodels.Conversation
	if err := s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
		c, err := s.conv.GetConversation(ctx, convID)
		conv = c
		return err
	}); err != nil {
		return nil, "", err
	}
	state, err := s.rdb.GetState(ctx, convID.String())
	if err != nil { // cache miss → fall back to PG and repopulate
		state = conv.State
		if state == "" {
			state = redisx.StateAI
		}
		_ = s.rdb.SetState(ctx, convID.String(), state)
	}
	return conv, state, nil
}

func (s *Service) handover(ctx context.Context, companyID uuid.UUID, conv *convmodels.Conversation) error {
	_ = s.rdb.SetState(ctx, conv.ID.String(), redisx.StateHuman)
	_ = s.rdb.Block(ctx, conv.ID.String(), passiveBlockTTL)
	return s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
		return s.conv.SetState(ctx, conv, convsvc.StateHuman)
	})
}

func (s *Service) deliver(ctx context.Context, companyID uuid.UUID, creds *repository.ChannelCreds, apiKey string, job jobs.InboundJob, conv *convmodels.Conversation, parts []string) {
	adapter, ok := s.adapters.For(creds.Type)
	if !ok {
		s.log.Error("no adapter for channel type", zap.String("type", creds.Type))
		return
	}
	out := channeladapter.Outbound{Instance: creds.InstanceName, APIKey: apiKey, Number: stripJID(job.RemoteJID)}
	s.log.Debug("pipeline: deliver", zap.String("instance", out.Instance), zap.String("number", out.Number), zap.Int("parts", len(parts)))
	for _, part := range parts {
		go func() { _ = adapter.SendPresence(ctx, out, channeladapter.PresenceComposing) }()
		msgID, err := adapter.SendText(ctx, out, part, sendDelayMs())
		if err != nil {
			s.log.Error("send text failed", zap.String("instance", out.Instance), zap.String("number", out.Number), zap.Error(err))
			continue
		}
		s.log.Debug("pipeline: message sent", zap.String("msg_id", msgID))
		_ = s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
			_, _, e := s.conv.AppendMessage(ctx, conv, convsvc.AppendInput{
				Direction: "outbound", SenderType: "ai", Content: part,
				ExternalMessageID: msgID, Status: "sent",
			})
			return e
		})
	}
}

func (s *Service) sendFallback(ctx context.Context, creds *repository.ChannelCreds, apiKey string, job jobs.InboundJob, msg string) {
	if msg == "" {
		return
	}
	if adapter, ok := s.adapters.For(creds.Type); ok {
		out := channeladapter.Outbound{Instance: creds.InstanceName, APIKey: apiKey, Number: stripJID(job.RemoteJID)}
		_, _ = adapter.SendText(ctx, out, msg, 0)
	}
}

func (s *Service) markRead(ctx context.Context, creds *repository.ChannelCreds, apiKey string, job jobs.InboundJob) {
	if adapter, ok := s.adapters.For(creds.Type); ok {
		out := channeladapter.Outbound{Instance: creds.InstanceName, APIKey: apiKey, Number: stripJID(job.RemoteJID)}
		_ = adapter.MarkRead(ctx, out, job.ExternalID)
	}
}

// passiveBlockTTL mirrors the webhook module's fromMe block duration.
const passiveBlockTTL = 30 * time.Minute

func toolset(handoverEnabled bool) []openai.Tool {
	if handoverEnabled {
		return []openai.Tool{transferToHumanTool}
	}
	return nil
}

func hasTransfer(calls []openai.ToolCall) bool {
	for _, c := range calls {
		if c.Function.Name == "transfer_to_human" {
			return true
		}
	}
	return false
}

func containsKeyword(text string, keywords []string) bool {
	lower := strings.ToLower(text)
	for _, k := range keywords {
		if k != "" && strings.Contains(lower, strings.ToLower(k)) {
			return true
		}
	}
	return false
}

func stripJID(jid string) string {
	if i := strings.IndexByte(jid, '@'); i >= 0 {
		jid = jid[:i]
	}
	if i := strings.IndexByte(jid, ':'); i >= 0 {
		jid = jid[:i]
	}
	return jid
}

func sendDelayMs() int {
	return minSendDelayMs + int(time.Now().UnixNano()%int64(maxSendDelayMs-minSendDelayMs))
}

func orInt(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}
