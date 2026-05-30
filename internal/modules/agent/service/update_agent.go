package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
)

// Update applies partial changes; system_prompt edits take effect immediately
// for new messages (RF-AG-03).
func (s *Service) Update(ctx context.Context, id uuid.UUID, in dto.UpdateAgentRequest) (*models.Agent, error) {
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
