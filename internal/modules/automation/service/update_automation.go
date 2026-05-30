package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
)

// Update applies partial changes and keeps channel.active_agent_id consistent.
func (s *Service) Update(ctx context.Context, id uuid.UUID, in dto.UpdateAutomationRequest) (*models.Automation, error) {
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
