// Package models holds the GORM entity for webhook audit/idempotency.
package models

import (
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
)

// WebhookEvent is a durable audit record of an inbound provider webhook.
type WebhookEvent struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CompanyID   *uuid.UUID `gorm:"type:uuid"`
	ChannelID   *uuid.UUID `gorm:"type:uuid"`
	EventType   string
	ExternalID  string
	Payload     database.JSONMap `gorm:"type:jsonb"`
	ProcessedAt *time.Time
	CreatedAt   time.Time
}

func (WebhookEvent) TableName() string { return "webhook_events" }
