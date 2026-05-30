// Package dto holds response payloads for the conversation module.
package dto

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
