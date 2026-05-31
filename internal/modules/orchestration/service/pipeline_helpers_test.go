package service

import (
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/pkg/openai"
	"github.com/stretchr/testify/require"
)

func TestToolset(t *testing.T) {
	require.Len(t, toolset(true), 1, "handover-enabled agents expose the transfer tool")
	require.Nil(t, toolset(false))
}

func TestHasTransfer(t *testing.T) {
	var transfer openai.ToolCall
	transfer.Function.Name = "transfer_to_human"
	var other openai.ToolCall
	other.Function.Name = "lookup"

	require.True(t, hasTransfer([]openai.ToolCall{other, transfer}))
	require.False(t, hasTransfer([]openai.ToolCall{other}))
	require.False(t, hasTransfer(nil))
}

func TestContainsKeyword(t *testing.T) {
	require.True(t, containsKeyword("I want a HUMAN please", []string{"human"}))
	require.True(t, containsKeyword("fala com atendente", []string{"gerente", "atendente"}))
	require.False(t, containsKeyword("all good", []string{"human"}))
	require.False(t, containsKeyword("anything", []string{""}), "empty keyword must never match")
}

func TestStripJID(t *testing.T) {
	require.Equal(t, "5511999", stripJID("5511999@s.whatsapp.net"))
	require.Equal(t, "5511999", stripJID("5511999:12@s.whatsapp.net"))
	require.Equal(t, "plain", stripJID("plain"))
}

func TestOrInt(t *testing.T) {
	require.Equal(t, 8, orInt(0, 8))
	require.Equal(t, 8, orInt(-3, 8))
	require.Equal(t, 5, orInt(5, 8))
}
