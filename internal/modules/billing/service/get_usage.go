package service

import (
	"context"
	"errors"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/billing/infra/repository"
	"github.com/google/uuid"
)

// UsageSummary is the current-period usage rollup for the panel dashboard.
type UsageSummary struct {
	Limits      *repository.Limits
	Totals      map[string]int64
	PeriodStart time.Time
	PeriodEnd   time.Time
}

// GetUsage aggregates the caller tenant's usage for the current billing period.
func (s *Service) GetUsage(ctx context.Context, companyID uuid.UUID) (*UsageSummary, error) {
	limits, err := s.Limits(ctx, companyID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			now := time.Now().UTC()
			return &UsageSummary{
				Limits:      &repository.Limits{},
				Totals:      map[string]int64{},
				PeriodStart: now,
				PeriodEnd:   now,
			}, nil
		}
		return nil, err
	}
	since := limits.PeriodStart
	if since.IsZero() {
		since = time.Now().UTC().AddDate(0, 0, -30)
	}
	rows, err := s.repo.SumUsageSince(ctx, since)
	if err != nil {
		return nil, err
	}
	totals := make(map[string]int64, len(rows))
	for _, r := range rows {
		totals[r.Kind] = r.Total
	}
	return &UsageSummary{
		Limits:      limits,
		Totals:      totals,
		PeriodStart: limits.PeriodStart,
		PeriodEnd:   limits.PeriodEnd,
	}, nil
}
