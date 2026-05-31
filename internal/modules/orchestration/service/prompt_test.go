package service

import (
	"strings"
	"testing"

	convmodels "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/pkg/openai"
	"github.com/stretchr/testify/require"
)

func TestBuildMessages_Structure(t *testing.T) {
	history := []convmodels.Message{
		{SenderType: "contact", Content: "hi"},
		{SenderType: "ai", Content: "hello"},
		{SenderType: "contact", Content: "   "}, // blank => skipped
	}
	msgs := buildMessages("You are a bot.", []string{"fact one"}, history, "what is up?")

	// system + 2 history (blank dropped) + user = 4
	require.Len(t, msgs, 4)
	require.Equal(t, openai.RoleSystem, msgs[0].Role)
	require.Contains(t, msgs[0].Content, "You are a bot.")
	require.Contains(t, msgs[0].Content, "fact one", "RAG chunk must be grounded into the system turn")
	require.Equal(t, openai.RoleUser, msgs[1].Role)      // contact "hi"
	require.Equal(t, openai.RoleAssistant, msgs[2].Role) // ai "hello"
	require.Equal(t, openai.RoleUser, msgs[3].Role)      // aggregated user message
	require.Equal(t, "what is up?", msgs[3].Content)
}

func TestBuildMessages_NoRAGKeepsSystemClean(t *testing.T) {
	msgs := buildMessages("Sys", nil, nil, "q")
	require.Len(t, msgs, 2)
	require.Equal(t, "Sys", msgs[0].Content)
}

func TestPrefixChunks(t *testing.T) {
	out := prefixChunks([]string{" a ", "b"})
	require.Equal(t, []string{"- a", "- b"}, out)
	require.True(t, strings.HasPrefix(out[0], "- "))
}
