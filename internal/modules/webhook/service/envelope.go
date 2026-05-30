package service

import (
	"encoding/json"
	"strings"
)

// Event types (normalized to UPPER_SNAKE).
const (
	EventMessagesUpsert   = "MESSAGES_UPSERT"
	EventConnectionUpdate = "CONNECTION_UPDATE"
	EventSendMessage      = "SEND_MESSAGE"
	EventMessagesUpdate   = "MESSAGES_UPDATE"
	EventQRCodeUpdated    = "QRCODE_UPDATED"
)

// envelope is the outer Evolution webhook payload.
type envelope struct {
	Event    string          `json:"event"`
	Instance string          `json:"instance"`
	Data     json.RawMessage `json:"data"`
}

// messageData is the MESSAGES_UPSERT data payload.
type messageData struct {
	Key struct {
		ID        string `json:"id"`
		RemoteJID string `json:"remoteJid"`
		FromMe    bool   `json:"fromMe"`
	} `json:"key"`
	PushName    string          `json:"pushName"`
	MessageType string          `json:"messageType"`
	Message     messageContent  `json:"message"`
	Status      string          `json:"status"`
	Raw         json.RawMessage `json:"-"`
}

type messageContent struct {
	Conversation        string `json:"conversation"`
	ExtendedTextMessage struct {
		Text string `json:"text"`
	} `json:"extendedTextMessage"`
}

// connectionData is the CONNECTION_UPDATE data payload.
type connectionData struct {
	State string `json:"state"`
}

// statusData is the SEND_MESSAGE / MESSAGES_UPDATE data payload.
type statusData struct {
	Key struct {
		ID string `json:"id"`
	} `json:"key"`
	Status string `json:"status"`
}

// normalizeEvent maps "messages.upsert" -> "MESSAGES_UPSERT".
func normalizeEvent(e string) string {
	return strings.ToUpper(strings.ReplaceAll(e, ".", "_"))
}

// text extracts the textual body of a message.
func (m messageData) text() string {
	if m.Message.Conversation != "" {
		return m.Message.Conversation
	}
	return m.Message.ExtendedTextMessage.Text
}

// mapConnectionState maps Evolution states to channel statuses.
func mapConnectionState(state string) string {
	switch state {
	case "open":
		return "connected"
	case "connecting":
		return "connecting"
	case "close":
		return "disconnected"
	default:
		return "error"
	}
}

// mapDeliveryStatus maps Evolution delivery states to message statuses.
func mapDeliveryStatus(s string) string {
	switch strings.ToUpper(s) {
	case "DELIVERY_ACK", "DELIVERED":
		return "delivered"
	case "READ", "PLAYED":
		return "read"
	case "SERVER_ACK", "SENT", "PENDING":
		return "sent"
	case "ERROR":
		return "failed"
	default:
		return "sent"
	}
}
