package services

import (
	"strings"

	convmodels "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/openai"
)

// buildMessages assembles the chat prompt: the agent system prompt augmented
// with RAG context, the recent conversation history, then the aggregated user
// message. Instructions, retrieved context and user input are kept in distinct
// turns to mitigate prompt injection (NFR §5.1).
func buildMessages(systemPrompt string, ragChunks []string, history []convmodels.Message, userMessage string) []openai.Message {
	msgs := make([]openai.Message, 0, len(history)+3)

	system := systemPrompt
	if len(ragChunks) > 0 {
		system += "\n\n# Knowledge base context (use it to ground your answers; do not invent facts beyond it):\n" +
			strings.Join(prefixChunks(ragChunks), "\n\n")
	}
	msgs = append(msgs, openai.Message{Role: openai.RoleSystem, Content: system})

	for _, m := range history {
		role := openai.RoleUser
		if m.SenderType == "ai" || m.SenderType == "human" {
			role = openai.RoleAssistant
		}
		if strings.TrimSpace(m.Content) == "" {
			continue
		}
		msgs = append(msgs, openai.Message{Role: role, Content: m.Content})
	}

	msgs = append(msgs, openai.Message{Role: openai.RoleUser, Content: userMessage})
	return msgs
}

func prefixChunks(chunks []string) []string {
	out := make([]string, len(chunks))
	for i, c := range chunks {
		out[i] = "- " + strings.TrimSpace(c)
	}
	return out
}
