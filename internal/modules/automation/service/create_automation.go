package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
)

// Create binds a channel to an agent. When active, it reflects the agent onto
// the channel and the DB rejects a second active automation for the channel.
func (s *Service) Create(ctx context.Context, in dto.CreateAutomationRequest) (*models.Automation, error) {
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
