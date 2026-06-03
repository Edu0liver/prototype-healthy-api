// Package service holds the agent use cases (one file per use case).
package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/repository"
	"github.com/google/uuid"
)

// Service implements the agent use cases.
type Service struct {
	repo Repository
	bill QuotaGuard
}

// New builds the agent service. Billing enforcement defaults to a no-op; the fx
// module wires the real guard via WithBilling so unit tests need no billing.
func New(repo Repository) *Service { return &Service{repo: repo, bill: noopQuota{}} }

// WithBilling installs the billing quota guard (production wiring).
func (s *Service) WithBilling(b QuotaGuard) *Service { s.bill = b; return s }

func (s *Service) get(ctx context.Context, id uuid.UUID) (*models.Agent, error) {
	a, err := s.repo.Get(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrAgentNotFound
	}
	return a, err
}

func uuidV7() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
func derefFloat(p *float64, def float64) float64 {
	if p == nil {
		return def
	}
	return *p
}
func derefInt(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}
func derefBool(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}
