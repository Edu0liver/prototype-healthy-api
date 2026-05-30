package dto

import "time"

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
