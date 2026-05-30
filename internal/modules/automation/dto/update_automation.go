package dto

// UpdateAutomationRequest updates mutable automation fields.
type UpdateAutomationRequest struct {
	AgentID         *string        `json:"agent_id" binding:"omitempty,uuid"`
	IsActive        *bool          `json:"is_active"`
	BusinessHours   map[string]any `json:"business_hours"`
	FallbackMessage *string        `json:"fallback_message"`
	DebounceSeconds *int           `json:"debounce_seconds" binding:"omitempty,gte=0,lte=60"`
}
