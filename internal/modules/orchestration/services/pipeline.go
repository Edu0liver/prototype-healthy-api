// Package services implements the orchestration pipeline: the core inbound
// processing described in PRD §2.6 plus the PROMPT.md extras (audio
// transcription, output humanization, handover triggers).
package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	convmodels "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	convsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/services"
	knowsvc "github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/services"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/orchestration/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/channeladapter"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/crypto"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/evolution"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/jobs"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/openai"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/redisx"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	bufferTTL      = 60 * time.Second
	lockTTL        = 90 * time.Second
	ragTopK        = 5
	historyLimit   = 10
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
	repo     *repositories.Repository
	conv     *convsvc.Service
	know     *knowsvc.Service
	evo      evolution.Client
	oa       openai.Client
	cipher   *crypto.Cipher
	adapters *channeladapter.Registry
}

// New builds the orchestration service.
func New(db *database.DB, rdb *redisx.Client, cfg *config.Config, log *zap.Logger,
	repo *repositories.Repository, conv *convsvc.Service, know *knowsvc.Service,
	evo evolution.Client, oa openai.Client, cipher *crypto.Cipher, adapters *channeladapter.Registry) *Service {
	return &Service{db: db, rdb: rdb, cfg: cfg, log: log, repo: repo, conv: conv,
		know: know, evo: evo, oa: oa, cipher: cipher, adapters: adapters}
}

// Process handles a single inbound job end-to-end.
func (s *Service) Process(ctx context.Context, job jobs.InboundJob) error {
	companyID, err := uuid.Parse(job.CompanyID)
	if err != nil {
		return err
	}
	channelID, _ := uuid.Parse(job.ChannelID)
	convID, _ := uuid.Parse(job.ConversationID)

	// Load channel creds + active agent (tenant scope).
	var creds *repositories.ChannelCreds
	var agentCfg *repositories.AgentConfig
	if err := s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
		creds, err = s.repo.LoadChannelCreds(ctx, channelID)
		if err != nil {
			return err
		}
		agentCfg, err = s.repo.LoadActiveAgent(ctx, channelID)
		if errors.Is(err, repositories.ErrNoActiveAgent) {
			return nil // no agent: nothing to do
		}
		return err
	}); err != nil {
		return err
	}
	if agentCfg == nil {
		return nil // no active agent on this channel — drop
	}
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

	// Serialize per-conversation processing (Redlock).
	lock, err := s.rdb.AcquireLock(ctx, redisx.LockKey(convID.String()), lockTTL)
	if errors.Is(err, redisx.ErrLockNotAcquired) {
		return nil // another worker owns this conversation; our fragment is buffered
	}
	if err != nil {
		return err
	}
	defer lock.Release(ctx)

	// Debounce window: wait for more fragments, then aggregate.
	debounce := time.Duration(orInt(agentCfg.DebounceSeconds, s.cfg.Worker.DebounceSeconds)) * time.Second
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(debounce):
	}
	fragments, err := s.rdb.DrainBuffer(ctx, convID.String())
	if err != nil {
		return err
	}
	if len(fragments) == 0 {
		return nil // already processed by a peer
	}
	aggregated := strings.TrimSpace(strings.Join(fragments, "\n"))
	if aggregated == "" {
		return nil
	}

	// Load conversation + resolve effective state (Redis, PG fallback).
	conv, state, err := s.loadConversation(ctx, companyID, convID)
	if err != nil {
		return err
	}
	if state == redisx.StateHuman {
		return nil // human is in control — IA stays silent (RF-HO-02)
	}
	if blocked, _ := s.rdb.IsBlocked(ctx, convID.String()); blocked {
		return nil
	}

	// Keyword-based handover (RF-HO-01).
	if agentCfg.HandoverEnabled && containsKeyword(aggregated, agentCfg.HandoverKeywords) {
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
		for _, h := range hits {
			ragChunks = append(ragChunks, h.Content)
		}
		history, _ = s.conv.RecentMessages(ctx, convID, historyLimit)
		return nil
	}); err != nil {
		return err
	}

	// OpenAI chat with function calling.
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
	if hasTransfer(result.ToolCalls) {
		return s.handover(ctx, companyID, conv)
	}

	// Humanize + send.
	parts := humanize(result.Content)
	if len(parts) == 0 {
		return nil
	}
	s.deliver(ctx, companyID, creds, apiKey, job, conv, parts)

	// Read receipt + mirror state.
	s.markRead(ctx, creds, apiKey, job)
	_ = s.rdb.SetState(ctx, convID.String(), redisx.StateAI)
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

func (s *Service) deliver(ctx context.Context, companyID uuid.UUID, creds *repositories.ChannelCreds, apiKey string, job jobs.InboundJob, conv *convmodels.Conversation, parts []string) {
	adapter, ok := s.adapters.For(creds.Type)
	if !ok {
		s.log.Error("no adapter for channel type", zap.String("type", creds.Type))
		return
	}
	out := channeladapter.Outbound{Instance: creds.InstanceName, APIKey: apiKey, Number: stripJID(job.RemoteJID)}
	for _, part := range parts {
		_ = adapter.SendPresence(ctx, out, channeladapter.PresenceComposing)
		msgID, err := adapter.SendText(ctx, out, part, sendDelayMs())
		if err != nil {
			s.log.Error("send text failed", zap.Error(err))
			continue
		}
		_ = s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
			_, _, e := s.conv.AppendMessage(ctx, conv, convsvc.AppendInput{
				Direction: "outbound", SenderType: "ai", Content: part,
				ExternalMessageID: msgID, Status: "sent",
			})
			return e
		})
	}
}

func (s *Service) sendFallback(ctx context.Context, creds *repositories.ChannelCreds, apiKey string, job jobs.InboundJob, msg string) {
	if msg == "" {
		return
	}
	if adapter, ok := s.adapters.For(creds.Type); ok {
		out := channeladapter.Outbound{Instance: creds.InstanceName, APIKey: apiKey, Number: stripJID(job.RemoteJID)}
		_, _ = adapter.SendText(ctx, out, msg, 0)
	}
}

func (s *Service) markRead(ctx context.Context, creds *repositories.ChannelCreds, apiKey string, job jobs.InboundJob) {
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
