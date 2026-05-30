// Package dto holds request/response payloads for the channel module.
package dto

// CreateChannelRequest creates a channel.
type CreateChannelRequest struct {
	Type   string `json:"type" binding:"required,oneof=whatsapp instagram"`
	Name   string `json:"name" binding:"omitempty"`
	Number string `json:"number" binding:"omitempty"` // external account / phone (E.164)
}
