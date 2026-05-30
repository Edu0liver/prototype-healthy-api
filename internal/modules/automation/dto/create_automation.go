// Package dto holds request/response payloads for the automation module.
package dto

// CreateAutomationRequest binds a channel to an agent.
type CreateAutomationRequest struct {
	ChannelID       string         `json:"channel_id" binding:"required,uuid"`
	AgentID         string         `json:"agent_id" binding:"required,uuid"`
	IsActive        *bool          `json:"is_active"`
	BusinessHours   map[string]any `json:"business_hours"`
	FallbackMessage string         `json:"fallback_message"`
	DebounceSeconds *int           `json:"debounce_seconds" binding:"omitempty,gte=0,lte=60"`
}
