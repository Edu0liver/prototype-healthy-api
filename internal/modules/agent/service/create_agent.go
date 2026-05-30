package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
)

// Create persists a new agent for the caller's tenant.
func (s *Service) Create(ctx context.Context, in dto.CreateAgentRequest) (*models.Agent, error) {
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
