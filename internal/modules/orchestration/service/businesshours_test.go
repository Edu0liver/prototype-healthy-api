package service

import (
	"testing"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
)

func TestWithinBusinessHours(t *testing.T) {
	// Mon 2024-01-01 10:00, Sat 2024-01-06 10:00 in UTC.
	mon10 := time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)
	mon20 := time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC)
	sat10 := time.Date(2024, 1, 6, 10, 0, 0, 0, time.UTC)

	cfg := database.JSONMap{
		"timezone": "UTC",
		"windows": map[string]any{
			"mon": []any{map[string]any{"start": "09:00", "end": "18:00"}},
		},
	}

	cases := []struct {
		name string
		raw  database.JSONMap
		now  time.Time
		want bool
	}{
		{"empty config is always open", nil, mon20, true},
		{"no windows key is always open", database.JSONMap{"timezone": "UTC"}, mon20, true},
		{"inside window", cfg, mon10, true},
		{"outside window same day", cfg, mon20, false},
		{"day not listed is closed", cfg, sat10, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := withinBusinessHours(c.raw, c.now); got != c.want {
				t.Fatalf("withinBusinessHours = %v, want %v", got, c.want)
			}
		})
	}
}

func TestWithinBusinessHours_MultipleWindows(t *testing.T) {
	cfg := database.JSONMap{
		"timezone": "UTC",
		"windows": map[string]any{
			"mon": []any{
				map[string]any{"start": "09:00", "end": "12:00"},
				map[string]any{"start": "13:00", "end": "18:00"},
			},
		},
	}
	lunch := time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC) // gap between windows
	afternoon := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)

	if withinBusinessHours(cfg, lunch) {
		t.Fatal("expected closed during lunch gap")
	}
	if !withinBusinessHours(cfg, afternoon) {
		t.Fatal("expected open in afternoon window")
	}
}

func TestWithinBusinessHours_Timezone(t *testing.T) {
	cfg := database.JSONMap{
		"timezone": "America/Sao_Paulo", // UTC-3
		"windows": map[string]any{
			"mon": []any{map[string]any{"start": "09:00", "end": "18:00"}},
		},
	}
	// 2024-01-01 11:00 UTC == 08:00 in São Paulo (before opening).
	utc11 := time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC)
	if withinBusinessHours(cfg, utc11) {
		t.Fatal("expected closed: 08:00 local is before 09:00 opening")
	}
	// 2024-01-01 13:00 UTC == 10:00 in São Paulo (open).
	utc13 := time.Date(2024, 1, 1, 13, 0, 0, 0, time.UTC)
	if !withinBusinessHours(cfg, utc13) {
		t.Fatal("expected open: 10:00 local is within hours")
	}
}
