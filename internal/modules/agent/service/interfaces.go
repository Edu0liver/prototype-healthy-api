package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/agent/infra/models"
	"github.com/google/uuid"
)

// Repository is the persistence contract consumed by the agent service.
type Repository interface {
	Create(ctx context.Context, a *models.Agent) error
	Update(ctx context.Context, a *models.Agent) error
	Get(ctx context.Context, id uuid.UUID) (*models.Agent, error)
	List(ctx context.Context) ([]models.Agent, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
