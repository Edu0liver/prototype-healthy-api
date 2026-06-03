package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/repository"
	"github.com/google/uuid"
)

const limitsCacheTTL = 5 * time.Minute

// UsageDecision is the outcome of a soft (usage) quota check.
type UsageDecision struct {
	Allowed        bool // proceed?
	Exceeded       bool // quota already reached
	OverageAllowed bool // plan opts into metered overage for this kind
}

// Limits returns the company's plan limits, cached in Redis (TTL). Reads via
// db.System because plans/subscriptions are not under RLS, so it works from any
// context (HTTP request, worker goroutine) without an ambient tenant tx.
func (s *Service) Limits(ctx context.Context, companyID uuid.UUID) (*repository.Limits, error) {
	cacheKey := "billing:limits:" + companyID.String()
	if s.rdb != nil {
		if raw, err := s.rdb.Get(ctx, cacheKey).Bytes(); err == nil {
			var l repository.Limits
			if json.Unmarshal(raw, &l) == nil {
				return &l, nil
			}
		}
	}
	var l *repository.Limits
	if err := s.db.System(ctx, func(ctx context.Context) error {
		var e error
		l, e = s.repo.LoadLimits(ctx, companyID)
		return e
	}); err != nil {
		return nil, err
	}
	if s.rdb != nil {
		if buf, err := json.Marshal(l); err == nil {
			_ = s.rdb.Set(ctx, cacheKey, buf, limitsCacheTTL).Err()
		}
	}
	return l, nil
}

// InvalidateLimits drops the cached limits (call on plan/subscription change).
func (s *Service) InvalidateLimits(ctx context.Context, companyID uuid.UUID) {
	if s.rdb == nil {
		return
	}
	_ = s.rdb.Del(ctx, "billing:limits:"+companyID.String()).Err()
}

// EnsureResource enforces a hard resource cap (channels/agents/knowledge_bases/
// seats) at create time. Returns ErrQuotaExceeded (HTTP 402) when at/over cap.
// Must run inside the caller's tenant tx (the count is tenant-scoped).
func (s *Service) EnsureResource(ctx context.Context, companyID uuid.UUID, resource string) error {
	limits, err := s.Limits(ctx, companyID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil // no subscription yet → fail open (do not block creates)
	}
	if err != nil {
		return err
	}
	table, max := resourceCap(limits, resource)
	if max <= 0 { // 0 = unlimited, or unknown resource → no cap
		return nil
	}
	count, err := s.repo.CountResource(ctx, table)
	if err != nil {
		return err
	}
	if count >= int64(max) {
		return ErrQuotaExceeded
	}
	return nil
}

// CheckUsage evaluates a soft usage limit (msgs/tokens/audio/storage) against
// the hot Redis counter for the current period. Default policy is hard-stop;
// a plan with overage pricing for the kind opts into continuing past quota.
func (s *Service) CheckUsage(ctx context.Context, companyID uuid.UUID, kind string) (UsageDecision, error) {
	limits, err := s.Limits(ctx, companyID)
	if err != nil {
		// No subscription / transient read error → fail open so the AI keeps
		// answering rather than going silent on a billing hiccup.
		return UsageDecision{Allowed: true}, nil
	}
	quota := quotaFor(limits, kind)
	if quota <= 0 { // unlimited / not gated
		return UsageDecision{Allowed: true}, nil
	}
	current := s.currentUsage(ctx, companyID, limits, kind)
	exceeded := current >= quota
	overage := overageFor(limits, kind) > 0
	return UsageDecision{Allowed: !exceeded || overage, Exceeded: exceeded, OverageAllowed: overage}, nil
}

func (s *Service) currentUsage(ctx context.Context, companyID uuid.UUID, limits *repository.Limits, kind string) int64 {
	if s.rdb == nil {
		return 0
	}
	n, err := s.rdb.Get(ctx, counterKey(companyID, periodKey(limits), kind)).Int64()
	if err != nil {
		return 0
	}
	return n
}

// ---- mappings -------------------------------------------------------------

func resourceCap(l *repository.Limits, resource string) (table string, max int) {
	switch resource {
	case "channels":
		return "channels", l.MaxChannels
	case "agents":
		return "agents", l.MaxAgents
	case "knowledge_bases":
		return "knowledge_bases", l.MaxKB
	case "seats":
		return "users", l.MaxSeats
	default:
		return "", 0
	}
}

func quotaFor(l *repository.Limits, kind string) int64 {
	switch kind {
	case models.KindAIMessage:
		return int64(l.QuotaAIMessages)
	case models.KindLLMTokens:
		return l.QuotaTokens
	case models.KindAudioMinutes:
		return int64(l.QuotaAudioMinutes)
	case models.KindStorageMB:
		return int64(l.QuotaStorageMB)
	default: // embedding_tokens etc.: metered but not gated
		return 0
	}
}

func overageFor(l *repository.Limits, kind string) int {
	switch kind {
	case models.KindAIMessage:
		return l.OveragePerMsgCents
	case models.KindLLMTokens:
		return l.OveragePer1kTokensCents
	default:
		return 0
	}
}

// periodKey derives a stable per-period bucket key from the subscription window.
func periodKey(l *repository.Limits) string {
	if l.PeriodStart.IsZero() {
		return time.Now().UTC().Format("20060102")
	}
	return l.PeriodStart.UTC().Format("20060102")
}

func counterKey(companyID uuid.UUID, period, kind string) string {
	return "usage:" + companyID.String() + ":" + period + ":" + kind
}
