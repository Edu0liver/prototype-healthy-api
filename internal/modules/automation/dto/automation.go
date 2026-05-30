package dto

// AutomationResponse describes an automation.
type AutomationResponse struct {
	ID              string `json:"id"`
	ChannelID       string `json:"channel_id"`
	AgentID         string `json:"agent_id"`
	IsActive        bool   `json:"is_active"`
	FallbackMessage string `json:"fallback_message"`
	DebounceSeconds int    `json:"debounce_seconds"`
}
