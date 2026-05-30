package dto

// AgentResponse describes an agent.
type AgentResponse struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	SystemPrompt     string   `json:"system_prompt"`
	Model            string   `json:"model"`
	Temperature      float64  `json:"temperature"`
	MaxOutputTokens  int      `json:"max_output_tokens"`
	HandoverEnabled  bool     `json:"handover_enabled"`
	HandoverKeywords []string `json:"handover_keywords"`
	Status           string   `json:"status"`
}
