package services

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/automation/infra/models"
	"github.com/google/uuid"
)

// Repository is the persistence contract consumed by the automation service.
type Repository interface {
	Create(ctx context.Context, a *models.Automation) error
	Update(ctx context.Context, a *models.Automation) error
	Get(ctx context.Context, id uuid.UUID) (*models.Automation, error)
	List(ctx context.Context) ([]models.Automation, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ChannelBelongsToTenant(ctx context.Context, channelID uuid.UUID) (bool, error)
	AgentBelongsToTenant(ctx context.Context, agentID uuid.UUID) (bool, error)
	SetChannelActiveAgent(ctx context.Context, channelID uuid.UUID, agentID *uuid.UUID) error
}
