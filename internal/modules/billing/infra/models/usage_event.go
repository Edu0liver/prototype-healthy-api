package models

import (
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
)

// Usage event kinds (mirror the SQL CHECK constraint).
const (
	KindAIMessage       = "ai_message"
	KindLLMTokens       = "llm_tokens"
	KindAudioMinutes    = "audio_minutes"
	KindStorageMB       = "storage_mb"
	KindEmbeddingTokens = "embedding_tokens"
)

// UsageEvent is one durable metering record (tenant-scoped, under RLS).
type UsageEvent struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID uuid.UUID `gorm:"type:uuid"`

	Kind     string
	Quantity int64

	ConversationID *uuid.UUID `gorm:"type:uuid"`
	AgentID        *uuid.UUID `gorm:"type:uuid"`
	Model          string

	Metadata  database.JSONMap `gorm:"type:jsonb"`
	CreatedAt time.Time
}

func (UsageEvent) TableName() string { return "usage_events" }
