// Package models holds the GORM entities for the channel module.
package models

import (
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/google/uuid"
)

// Channel is a messaging connection (WhatsApp via Evolution, or Instagram).
type Channel struct {
	ID                    uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID             uuid.UUID `gorm:"type:uuid"`
	Type                  string    // whatsapp | instagram
	Name                  string
	EvolutionInstanceName string
	EvolutionInstanceID   string
	EvolutionAPIKeyEnc    string `gorm:"column:evolution_apikey_enc"`
	ExternalAccountID     string
	Status                string           // disconnected | connecting | connected | error
	ActiveAgentID         *uuid.UUID       `gorm:"type:uuid"`
	Metadata              database.JSONMap `gorm:"type:jsonb"`
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (Channel) TableName() string { return "channels" }
