package service

import (
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/models"
)

func TestMapStatus(t *testing.T) {
	cases := map[string]string{
		"active":             models.StatusActive,
		"trialing":           models.StatusTrialing,
		"past_due":           models.StatusPastDue,
		"canceled":           models.StatusCanceled,
		"unpaid":             models.StatusPastDue,
		"incomplete":         models.StatusSuspended,
		"incomplete_expired": models.StatusSuspended,
		"paused":             models.StatusSuspended,
		"weird_unknown":      models.StatusActive,
	}
	for in, want := range cases {
		if got := mapStatus(in); got != want {
			t.Errorf("mapStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if firstNonEmpty("", "", "x", "y") != "x" {
		t.Fatal("expected first non-empty 'x'")
	}
	if firstNonEmpty("", "") != "" {
		t.Fatal("expected empty when all empty")
	}
}
