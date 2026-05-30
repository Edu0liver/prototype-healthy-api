package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/google/uuid"
)

// Repository is the persistence contract consumed by the channel service.
type Repository interface {
	Create(ctx context.Context, c *models.Channel) error
	Update(ctx context.Context, c *models.Channel) error
	Get(ctx context.Context, id uuid.UUID) (*models.Channel, error)
	List(ctx context.Context) ([]models.Channel, error)
}
