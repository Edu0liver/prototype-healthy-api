package dto

// ChannelResponse describes a channel.
type ChannelResponse struct {
	ID                string  `json:"id"`
	Type              string  `json:"type"`
	Name              string  `json:"name"`
	Status            string  `json:"status"`
	ExternalAccountID string  `json:"external_account_id"`
	InstanceName      string  `json:"instance_name,omitempty"`
	ActiveAgentID     *string `json:"active_agent_id,omitempty"`
}
