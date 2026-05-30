// Package services holds the agent use cases.
package services

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/dtos"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/google/uuid"
)

// Service implements the agent use cases.
type Service struct {
	repo Repository
}

// New builds the agent service.
func New(repo Repository) *Service { return &Service{repo: repo} }

// Create persists a new agent for the caller's tenant.
func (s *Service) Create(ctx context.Context, in dtos.CreateAgentRequest) (*models.Agent, error) {
	a := &models.Agent{
		ID:               uuidV7(),
		CompanyID:        appctx.CompanyID(ctx),
		Name:             in.Name,
		SystemPrompt:     in.SystemPrompt,
		Model:            orDefault(in.Model, "gpt-4o-mini"),
		Temperature:      derefFloat(in.Temperature, 0.7),
		MaxOutputTokens:  derefInt(in.MaxOutputTokens, 1024),
		HandoverEnabled:  derefBool(in.HandoverEnabled, true),
		HandoverKeywords: database.JSONStringArray(in.HandoverKeywords),
		Status:           orDefault(in.Status, "draft"),
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Update applies partial changes; system_prompt edits take effect immediately
// for new messages (RF-AG-03).
func (s *Service) Update(ctx context.Context, id uuid.UUID, in dtos.UpdateAgentRequest) (*models.Agent, error) {
	a, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		a.Name = *in.Name
	}
	if in.SystemPrompt != nil {
		a.SystemPrompt = *in.SystemPrompt
	}
	if in.Model != nil {
		a.Model = *in.Model
	}
	if in.Temperature != nil {
		a.Temperature = *in.Temperature
	}
	if in.MaxOutputTokens != nil {
		a.MaxOutputTokens = *in.MaxOutputTokens
	}
	if in.HandoverEnabled != nil {
		a.HandoverEnabled = *in.HandoverEnabled
	}
	if in.HandoverKeywords != nil {
		a.HandoverKeywords = database.JSONStringArray(in.HandoverKeywords)
	}
	if in.Status != nil {
		a.Status = *in.Status
	}
	if err := s.repo.Update(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Get returns an agent by id.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*models.Agent, error) {
	return s.get(ctx, id)
}

// List returns all agents in the tenant.
func (s *Service) List(ctx context.Context) ([]models.Agent, error) {
	return s.repo.List(ctx)
}

// Delete removes an agent.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return ErrAgentNotFound
		}
		return err
	}
	return nil
}

func (s *Service) get(ctx context.Context, id uuid.UUID) (*models.Agent, error) {
	a, err := s.repo.Get(ctx, id)
	if errors.Is(err, repositories.ErrNotFound) {
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
