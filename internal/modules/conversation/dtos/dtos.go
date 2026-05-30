// Package dtos holds response payloads for the conversation module.
package dtos

import "time"

// ConversationResponse describes a conversation row.
type ConversationResponse struct {
	ID             string     `json:"id"`
	ChannelID      string     `json:"channel_id"`
	ContactID      string     `json:"contact_id"`
	AgentID        *string    `json:"agent_id,omitempty"`
	State          string     `json:"state"`
	AssignedUserID *string    `json:"assigned_user_id,omitempty"`
	LastMessageAt  *time.Time `json:"last_message_at,omitempty"`
	OpenedAt       time.Time  `json:"opened_at"`
	ClosedAt       *time.Time `json:"closed_at,omitempty"`
}

// MessageResponse describes a message row.
type MessageResponse struct {
	ID                string         `json:"id"`
	Direction         string         `json:"direction"`
	SenderType        string         `json:"sender_type"`
	Content           string         `json:"content"`
	Media             map[string]any `json:"media,omitempty"`
	ExternalMessageID string         `json:"external_message_id,omitempty"`
	Status            string         `json:"status"`
	CreatedAt         time.Time      `json:"created_at"`
}
