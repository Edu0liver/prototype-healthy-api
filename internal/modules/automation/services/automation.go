// Package services holds the automation use cases (channel↔agent binding with
// the "one active automation per channel" invariant).
package services

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/dtos"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/google/uuid"
)

// Service implements automation use cases.
type Service struct {
	repo Repository
	db   *database.DB
}

// New builds the automation service.
func New(repo Repository, db *database.DB) *Service { return &Service{repo: repo, db: db} }

// Create binds a channel to an agent. When active, it reflects the agent onto
// the channel and the DB rejects a second active automation for the channel.
func (s *Service) Create(ctx context.Context, in dtos.CreateAutomationRequest) (*models.Automation, error) {
	channelID, err := uuid.Parse(in.ChannelID)
	if err != nil {
		return nil, ErrChannelNotFound
	}
	agentID, err := uuid.Parse(in.AgentID)
	if err != nil {
		return nil, ErrAgentNotFound
	}
	if err := s.validateRefs(ctx, channelID, agentID); err != nil {
		return nil, err
	}

	a := &models.Automation{
		ID:              uuidV7(),
		CompanyID:       appctx.CompanyID(ctx),
		ChannelID:       channelID,
		AgentID:         agentID,
		IsActive:        derefBool(in.IsActive, true),
		BusinessHours:   database.JSONMap(in.BusinessHours),
		FallbackMessage: in.FallbackMessage,
		DebounceSeconds: derefInt(in.DebounceSeconds, 8),
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, mapRepoErr(err)
	}
	if a.IsActive {
		if err := s.repo.SetChannelActiveAgent(ctx, channelID, &agentID); err != nil {
			return nil, err
		}
	}
	return a, nil
}

// Update applies partial changes and keeps channel.active_agent_id consistent.
func (s *Service) Update(ctx context.Context, id uuid.UUID, in dtos.UpdateAutomationRequest) (*models.Automation, error) {
	a, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}
	wasActive := a.IsActive

	if in.AgentID != nil {
		agentID, err := uuid.Parse(*in.AgentID)
		if err != nil {
			return nil, ErrAgentNotFound
		}
		ok, err := s.repo.AgentBelongsToTenant(ctx, agentID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrAgentNotFound
		}
		a.AgentID = agentID
	}
	if in.IsActive != nil {
		a.IsActive = *in.IsActive
	}
	if in.BusinessHours != nil {
		a.BusinessHours = database.JSONMap(in.BusinessHours)
	}
	if in.FallbackMessage != nil {
		a.FallbackMessage = *in.FallbackMessage
	}
	if in.DebounceSeconds != nil {
		a.DebounceSeconds = *in.DebounceSeconds
	}
	if err := s.repo.Update(ctx, a); err != nil {
		return nil, mapRepoErr(err)
	}

	// Reflect active-agent changes onto the channel.
	switch {
	case a.IsActive:
		if err := s.repo.SetChannelActiveAgent(ctx, a.ChannelID, &a.AgentID); err != nil {
			return nil, err
		}
	case wasActive && !a.IsActive:
		if err := s.repo.SetChannelActiveAgent(ctx, a.ChannelID, nil); err != nil {
			return nil, err
		}
	}
	return a, nil
}

// Get returns an automation by id.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*models.Automation, error) {
	return s.get(ctx, id)
}

// List returns all automations in the tenant.
func (s *Service) List(ctx context.Context) ([]models.Automation, error) {
	return s.repo.List(ctx)
}

// Delete removes an automation and clears the channel's active agent if needed.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	a, err := s.get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if a.IsActive {
		return s.repo.SetChannelActiveAgent(ctx, a.ChannelID, nil)
	}
	return nil
}

func (s *Service) validateRefs(ctx context.Context, channelID, agentID uuid.UUID) error {
	ok, err := s.repo.ChannelBelongsToTenant(ctx, channelID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrChannelNotFound
	}
	ok, err = s.repo.AgentBelongsToTenant(ctx, agentID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrAgentNotFound
	}
	return nil
}

func (s *Service) get(ctx context.Context, id uuid.UUID) (*models.Automation, error) {
	a, err := s.repo.Get(ctx, id)
	if errors.Is(err, repositories.ErrNotFound) {
		return nil, ErrAutomationNotFound
	}
	return a, err
}

func mapRepoErr(err error) error {
	if errors.Is(err, repositories.ErrActiveExists) {
		return ErrActiveExists
	}
	return err
}

func uuidV7() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}

func derefBool(p *bool, def bool) bool {
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
