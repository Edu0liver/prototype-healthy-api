package service

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Event is a single usage observation handed to the metering pipeline.
type Event struct {
	CompanyID      uuid.UUID
	Kind           string
	Quantity       int64
	ConversationID *uuid.UUID
	AgentID        *uuid.UUID
	Model          string
}

// Record meters one usage event: it bumps the hot Redis counter (authoritative
// for quota checks) and enqueues the event on the durable outbox stream for the
// meter worker to persist. It is best-effort and never returns an error, so
// callers can fire it in a detached goroutine off the hot path. Pass a
// non-request-scoped context (e.g. context.WithoutCancel).
func (s *Service) Record(ctx context.Context, e Event) {
	if e.CompanyID == uuid.Nil || e.Quantity <= 0 || s.rdb == nil {
		return
	}

	// Hot counter (best-effort): increment and ensure a TTL to the period end.
	limits, _ := s.Limits(ctx, e.CompanyID)
	period := time.Now().UTC().Format("20060102")
	ttl := 32 * 24 * time.Hour
	if limits != nil {
		period = periodKey(limits)
		if d := time.Until(limits.PeriodEnd); d > 0 {
			ttl = d
		}
	}
	key := counterKey(e.CompanyID, period, e.Kind)
	if err := s.rdb.IncrBy(ctx, key, e.Quantity).Err(); err != nil {
		s.log.Warn("billing: counter incr failed", zap.Error(err))
	} else {
		_ = s.rdb.ExpireNX(ctx, key, ttl).Err()
	}

	// Durable outbox for the per-period ledger.
	if _, err := s.rdb.Enqueue(ctx, usageStream, e.toMap()); err != nil {
		s.log.Warn("billing: usage enqueue failed", zap.Error(err))
	}
}

func (e Event) toMap() map[string]any {
	m := map[string]any{
		"company_id": e.CompanyID.String(),
		"kind":       e.Kind,
		"quantity":   strconv.FormatInt(e.Quantity, 10),
		"model":      e.Model,
	}
	if e.ConversationID != nil {
		m["conversation_id"] = e.ConversationID.String()
	}
	if e.AgentID != nil {
		m["agent_id"] = e.AgentID.String()
	}
	return m
}

func eventFromMap(v map[string]any) Event {
	get := func(k string) string {
		if s, ok := v[k].(string); ok {
			return s
		}
		return ""
	}
	e := Event{Kind: get("kind"), Model: get("model")}
	e.CompanyID, _ = uuid.Parse(get("company_id"))
	e.Quantity, _ = strconv.ParseInt(get("quantity"), 10, 64)
	if s := get("conversation_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			e.ConversationID = &id
		}
	}
	if s := get("agent_id"); s != "" {
		if id, err := uuid.Parse(s); err == nil {
			e.AgentID = &id
		}
	}
	return e
}
