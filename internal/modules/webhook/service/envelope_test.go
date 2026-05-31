package service

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeEvent(t *testing.T) {
	require.Equal(t, "MESSAGES_UPSERT", normalizeEvent("messages.upsert"))
	require.Equal(t, "CONNECTION_UPDATE", normalizeEvent("connection.update"))
	require.Equal(t, "MESSAGES_UPSERT", normalizeEvent("MESSAGES_UPSERT"))
}

func TestMessageData_Text(t *testing.T) {
	// Plain conversation wins.
	var d messageData
	require.NoError(t, json.Unmarshal([]byte(`{"message":{"conversation":"hi there"}}`), &d))
	require.Equal(t, "hi there", d.text())

	// Falls back to extendedTextMessage when conversation is empty.
	var d2 messageData
	require.NoError(t, json.Unmarshal([]byte(`{"message":{"extendedTextMessage":{"text":"quoted reply"}}}`), &d2))
	require.Equal(t, "quoted reply", d2.text())

	// No text => empty.
	require.Equal(t, "", messageData{}.text())
}

func TestMapConnectionState(t *testing.T) {
	cases := map[string]string{
		"open":       "connected",
		"connecting": "connecting",
		"close":      "disconnected",
		"unknown":    "error",
	}
	for in, want := range cases {
		require.Equal(t, want, mapConnectionState(in), in)
	}
}

func TestMapDeliveryStatus(t *testing.T) {
	cases := map[string]string{
		"DELIVERY_ACK": "delivered",
		"delivered":    "delivered",
		"READ":         "read",
		"PLAYED":       "read",
		"SERVER_ACK":   "sent",
		"SENT":         "sent",
		"PENDING":      "sent",
		"ERROR":        "failed",
		"whatever":     "sent",
	}
	for in, want := range cases {
		require.Equal(t, want, mapDeliveryStatus(in), in)
	}
}

func TestEnvelope_Unmarshal(t *testing.T) {
	var env envelope
	require.NoError(t, json.Unmarshal([]byte(`{"event":"messages.upsert","instance":"lumia-x","data":{"key":{"id":"abc"}}}`), &env))
	require.Equal(t, "lumia-x", env.Instance)
	require.Equal(t, "messages.upsert", env.Event)
	require.NotEmpty(t, env.Data)
}
