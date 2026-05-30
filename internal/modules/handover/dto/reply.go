// Package dto holds request payloads for the handover module.
package dto

// ReplyRequest carries the operator's message.
type ReplyRequest struct {
	Content string `json:"content" binding:"required"`
}
