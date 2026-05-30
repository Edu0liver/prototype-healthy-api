// Package dtos holds request/response payloads for the automation module.
package dtos

// CreateAutomationRequest binds a channel to an agent.
type CreateAutomationRequest struct {
	ChannelID       string         `json:"channel_id" binding:"required,uuid"`
	AgentID         string         `json:"agent_id" binding:"required,uuid"`
	IsActive        *bool          `json:"is_active"`
	BusinessHours   map[string]any `json:"business_hours"`
	FallbackMessage string         `json:"fallback_message"`
	DebounceSeconds *int           `json:"debounce_seconds" binding:"omitempty,gte=0,lte=60"`
}

// UpdateAutomationRequest updates mutable automation fields.
type UpdateAutomationRequest struct {
	AgentID         *string        `json:"agent_id" binding:"omitempty,uuid"`
	IsActive        *bool          `json:"is_active"`
	BusinessHours   map[string]any `json:"business_hours"`
	FallbackMessage *string        `json:"fallback_message"`
	DebounceSeconds *int           `json:"debounce_seconds" binding:"omitempty,gte=0,lte=60"`
}

// AutomationResponse describes an automation.
type AutomationResponse struct {
	ID              string `json:"id"`
	ChannelID       string `json:"channel_id"`
	AgentID         string `json:"agent_id"`
	IsActive        bool   `json:"is_active"`
	FallbackMessage string `json:"fallback_message"`
	DebounceSeconds int    `json:"debounce_seconds"`
}
