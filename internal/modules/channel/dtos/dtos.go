// Package dtos holds request/response payloads for the channel module.
package dtos

// CreateChannelRequest creates a channel.
type CreateChannelRequest struct {
	Type   string `json:"type" binding:"required,oneof=whatsapp instagram"`
	Name   string `json:"name" binding:"omitempty"`
	Number string `json:"number" binding:"omitempty"` // external account / phone (E.164)
}

// ConnectRequest selects the connection method.
type ConnectRequest struct {
	Method string `json:"method" binding:"omitempty,oneof=qr pairing"`
	Number string `json:"number" binding:"omitempty"`
}

// ConnectResponse carries QR / pairing data.
type ConnectResponse struct {
	QRCode      string `json:"qr_code,omitempty"`
	PairingCode string `json:"pairing_code,omitempty"`
	State       string `json:"state"`
}

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
