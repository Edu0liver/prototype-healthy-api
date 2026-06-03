// Package models holds the GORM entities for the billing module.
package models

import (
	"time"

	"github.com/google/uuid"
)

// Plan is a global catalogue tier (no company_id, not under RLS). Quota and cap
// columns use 0 to mean "unlimited"; overage columns use 0 to mean "disabled"
// (hard-stop).
type Plan struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	Code       string
	Name       string
	PriceCents int
	Currency   string

	QuotaAIMessages   int
	QuotaTokens       int64
	QuotaAudioMinutes int
	QuotaStorageMB    int

	MaxChannels int
	MaxAgents   int
	MaxKB       int
	MaxSeats    int

	OveragePerMsgCents      int
	OveragePer1kTokensCents int

	StripePriceID   string
	StripeProductID string

	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Plan) TableName() string { return "plans" }
