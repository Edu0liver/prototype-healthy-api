package service

import (
	"context"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
	"github.com/google/uuid"
)

func TestEventRoundTrip(t *testing.T) {
	conv := uuid.New()
	agent := uuid.New()
	in := Event{
		CompanyID:      uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Kind:           models.KindLLMTokens,
		Quantity:       1234,
		ConversationID: &conv,
		AgentID:        &agent,
		Model:          "gpt-4o",
	}
	out := eventFromMap(in.toMap())

	if out.CompanyID != in.CompanyID || out.Kind != in.Kind || out.Quantity != in.Quantity || out.Model != in.Model {
		t.Errorf("scalar round trip mismatch: %+v != %+v", out, in)
	}
	if out.ConversationID == nil || *out.ConversationID != conv {
		t.Error("conversation id did not round trip")
	}
	if out.AgentID == nil || *out.AgentID != agent {
		t.Error("agent id did not round trip")
	}
}

func TestRecord_NilRedisIsNoop(t *testing.T) {
	// With no Redis client wired, Record must be a safe no-op (degraded mode),
	// never panicking on the hot path.
	s := &Service{}
	s.Record(context.Background(), Event{CompanyID: uuid.New(), Kind: models.KindAIMessage, Quantity: 1})
}
