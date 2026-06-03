package service

import (
	"testing"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/repository"
)

func testLimits() *repository.Limits {
	return &repository.Limits{
		QuotaAIMessages:         100,
		QuotaTokens:             50000,
		QuotaAudioMinutes:       10,
		QuotaStorageMB:          50,
		MaxChannels:             1,
		MaxAgents:               2,
		MaxKB:                   3,
		MaxSeats:                4,
		OveragePerMsgCents:      0,
		OveragePer1kTokensCents: 5,
	}
}

func TestQuotaFor(t *testing.T) {
	l := testLimits()
	cases := map[string]int64{
		models.KindAIMessage:       100,
		models.KindLLMTokens:       50000,
		models.KindAudioMinutes:    10,
		models.KindStorageMB:       50,
		models.KindEmbeddingTokens: 0, // metered but not gated
	}
	for kind, want := range cases {
		if got := quotaFor(l, kind); got != want {
			t.Errorf("quotaFor(%s) = %d, want %d", kind, got, want)
		}
	}
}

func TestOverageFor(t *testing.T) {
	l := testLimits()
	if overageFor(l, models.KindAIMessage) != 0 {
		t.Error("ai_message overage should be 0 (hard-stop)")
	}
	if overageFor(l, models.KindLLMTokens) != 5 {
		t.Error("llm_tokens overage should be 5")
	}
	if overageFor(l, models.KindStorageMB) != 0 {
		t.Error("storage overage should be 0")
	}
}

func TestResourceCap(t *testing.T) {
	l := testLimits()
	cases := []struct {
		resource  string
		wantTable string
		wantMax   int
	}{
		{"channels", "channels", 1},
		{"agents", "agents", 2},
		{"knowledge_bases", "knowledge_bases", 3},
		{"seats", "users", 4},
		{"unknown", "", 0},
	}
	for _, c := range cases {
		table, max := resourceCap(l, c.resource)
		if table != c.wantTable || max != c.wantMax {
			t.Errorf("resourceCap(%s) = (%q,%d), want (%q,%d)", c.resource, table, max, c.wantTable, c.wantMax)
		}
	}
}

func TestPeriodKey(t *testing.T) {
	l := &repository.Limits{PeriodStart: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)}
	if got := periodKey(l); got != "20260601" {
		t.Errorf("periodKey = %q, want 20260601", got)
	}
	// Zero start falls back to today's date.
	if got := periodKey(&repository.Limits{}); got != time.Now().UTC().Format("20060102") {
		t.Errorf("periodKey zero = %q, want today", got)
	}
}
