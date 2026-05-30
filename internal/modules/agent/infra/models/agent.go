// Package models holds the GORM entities for the agent module.
package models

import (
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
)

// Agent is an AI persona: system prompt, model params and handover config.
type Agent struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID        uuid.UUID `gorm:"type:uuid"`
	Name             string
	SystemPrompt     string
	Model            string
	Temperature      float64 `gorm:"type:numeric(3,2)"`
	MaxOutputTokens  int
	HandoverEnabled  bool
	HandoverKeywords database.JSONStringArray `gorm:"type:jsonb"`
	Status           string                   // active | draft
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (Agent) TableName() string { return "agents" }
