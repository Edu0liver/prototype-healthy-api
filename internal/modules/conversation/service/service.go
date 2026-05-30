// Package service holds the conversation use cases: pipeline persistence
// helpers (used by webhook + orchestration) and panel history/listing.
package service

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/events"
	"github.com/google/uuid"
)

// Conversation states.
const (
	StateAI     = "ai"
	StateHuman  = "human"
	StateClosed = "closed"
)

// Service implements conversation use cases.
type Service struct {
	repo   Repository
	events *events.Publisher
}

// New builds the conversation service.
func New(repo Repository, pub *events.Publisher) *Service {
	return &Service{repo: repo, events: pub}
}

func uuidV7() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}
