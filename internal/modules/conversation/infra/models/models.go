// Package models holds the GORM entities for the conversation module.
package models

import (
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/google/uuid"
)

// Contact is an end-user reachable on a channel.
type Contact struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID     uuid.UUID `gorm:"type:uuid"`
	ChannelID     uuid.UUID `gorm:"type:uuid"`
	RemoteJID     string    `gorm:"column:remote_jid"`
	PushName      string
	ProfilePicURL string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Contact) TableName() string { return "contacts" }

// Conversation is a thread between a contact and the platform on a channel.
type Conversation struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	CompanyID      uuid.UUID  `gorm:"type:uuid"`
	ChannelID      uuid.UUID  `gorm:"type:uuid"`
	ContactID      uuid.UUID  `gorm:"type:uuid"`
	AgentID        *uuid.UUID `gorm:"type:uuid"`
	State          string     // ai | human | closed
	AssignedUserID *uuid.UUID `gorm:"type:uuid"`
	LastMessageAt  *time.Time
	OpenedAt       time.Time
	ClosedAt       *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (Conversation) TableName() string { return "conversations" }

// Message is a single inbound/outbound message in a conversation.
type Message struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID         uuid.UUID `gorm:"type:uuid"`
	ConversationID    uuid.UUID `gorm:"type:uuid"`
	Direction         string    // inbound | outbound
	SenderType        string    // contact | ai | human
	Content           string
	Media             database.JSONMap `gorm:"type:jsonb"`
	ExternalMessageID string
	Status            string // received | sent | delivered | read | failed
	CreatedAt         time.Time
}

func (Message) TableName() string { return "messages" }
