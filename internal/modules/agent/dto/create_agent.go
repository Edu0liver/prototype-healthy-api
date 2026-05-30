// Package dto holds request/response payloads for the agent module.
package dto

// CreateAgentRequest creates an agent.
type CreateAgentRequest struct {
	Name             string   `json:"name" binding:"required,min=2"`
	SystemPrompt     string   `json:"system_prompt" binding:"required"`
	Model            string   `json:"model" binding:"omitempty"`
	Temperature      *float64 `json:"temperature" binding:"omitempty,gte=0,lte=2"`
	MaxOutputTokens  *int     `json:"max_output_tokens" binding:"omitempty,gt=0"`
	HandoverEnabled  *bool    `json:"handover_enabled"`
	HandoverKeywords []string `json:"handover_keywords"`
	Status           string   `json:"status" binding:"omitempty,oneof=active draft"`
}
