// Package service holds the automation use cases (one file per use case).
package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
)

// Service implements automation use cases.
type Service struct {
	repo Repository
	db   *database.DB
}

// New builds the automation service.
func New(repo Repository, db *database.DB) *Service { return &Service{repo: repo, db: db} }

func (s *Service) get(ctx context.Context, id uuid.UUID) (*models.Automation, error) {
	a, err := s.repo.Get(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrAutomationNotFound
	}
	return a, err
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

func mapRepoErr(err error) error {
	if errors.Is(err, repository.ErrActiveExists) {
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
