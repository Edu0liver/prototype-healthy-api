package models

import (
	"time"

	"github.com/google/uuid"
)

// Subscription statuses.
const (
	StatusTrialing  = "trialing"
	StatusActive    = "active"
	StatusPastDue   = "past_due"
	StatusCanceled  = "canceled"
	StatusSuspended = "suspended"
)

// Subscription is one company's plan binding. System-scoped (not under RLS);
// the gateway webhook writes it without tenant context.
type Subscription struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID uuid.UUID `gorm:"type:uuid"`
	PlanID    uuid.UUID `gorm:"type:uuid"`

	Status       string
	BillingCycle string

	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAtPeriodEnd  bool

	StripeCustomerID     string
	StripeSubscriptionID string

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Subscription) TableName() string { return "subscriptions" }
