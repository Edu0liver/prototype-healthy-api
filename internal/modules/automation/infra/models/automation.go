// Package models holds the GORM entities for the automation module.
package models

import (
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/google/uuid"
)

// Automation binds a channel to an agent with operating rules. At most one
// active automation per channel (partial unique index).
type Automation struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID       uuid.UUID `gorm:"type:uuid"`
	ChannelID       uuid.UUID `gorm:"type:uuid"`
	AgentID         uuid.UUID `gorm:"type:uuid"`
	IsActive        bool
	BusinessHours   database.JSONMap `gorm:"type:jsonb"`
	FallbackMessage string
	DebounceSeconds int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (Automation) TableName() string { return "automations" }
