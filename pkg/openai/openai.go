// Package openai is a thin client over the OpenAI HTTP API covering the three
// capabilities the platform needs: chat completions (with function calling),
// embeddings for RAG, and Whisper audio transcription.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/pkg/resilience"
)

// Role constants for chat messages.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// Message is a single chat message.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Name       string     `json:"name,omitempty"`
}

// Tool declares a callable function to the model.
type Tool struct {
	Type     string      `json:"type"` // "function"
	Function FunctionDef `json:"function"`
}

// FunctionDef is a function-calling tool definition.
type FunctionDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

// ToolCall is a function invocation requested by the model.
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// ChatRequest is a chat completion request.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []Tool    `json:"tools,omitempty"`
	Temperature float32   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// Usage reports the token accounting returned by the OpenAI API. It feeds the
// billing/metering pipeline (llm_tokens). Zero values mean the provider did not
// report usage (e.g. a mock or an older endpoint).
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatResult is the normalized chat response.
type ChatResult struct {
	Content   string
	ToolCalls []ToolCall
	Usage     Usage
}

// Client is the OpenAI API contract (interface for mocking).
type Client interface {
	Chat(ctx context.Context, req ChatRequest) (*ChatResult, error)
	Embed(ctx context.Context, inputs []string) ([][]float32, error)
	Transcribe(ctx context.Context, audio []byte, filename string) (string, error)
}

// Config configures the OpenAI client.
type Config struct {
	APIKey         string
	BaseURL        string
	EmbeddingModel string
	WhisperModel   string
	Timeout        time.Duration
}

// HTTPClient is the live implementation.
type HTTPClient struct {
	cfg  Config
	http *http.Client
}

// New builds the live OpenAI client.
func New(cfg Config) *HTTPClient {
	return &HTTPClient{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

func (c *HTTPClient) post(ctx context.Context, path string, body any, out any) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return c.send(ctx, func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+path, bytes.NewReader(buf))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	}, out)
}

// send executes the request with retry/backoff. Network errors and 429/5xx are
// retried; 4xx are permanent. build is called per attempt to get a fresh body.
func (c *HTTPClient) send(ctx context.Context, build func() (*http.Request, error), out any) error {
	return resilience.Do(ctx, resilience.DefaultRetry(), func() error {
		req, err := build()
		if err != nil {
			return resilience.Permanent(err)
		}
		resp, err := c.http.Do(req)
		if err != nil {
			return err // network error: retry
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			return fmt.Errorf("openai: %s: %s", resp.Status, string(data))
		}
		if resp.StatusCode >= 300 {
			return resilience.Permanent(fmt.Errorf("openai: %s: %s", resp.Status, string(data)))
		}
		if out != nil {
			return json.Unmarshal(data, out)
		}
		return nil
	})
}

// Chat performs a chat completion.
func (c *HTTPClient) Chat(ctx context.Context, req ChatRequest) (*ChatResult, error) {
	var resp struct {
		Choices []struct {
			Message struct {
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Usage Usage `json:"usage"`
	}
	if err := c.post(ctx, "/chat/completions", req, &resp); err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("openai: empty chat response")
	}
	m := resp.Choices[0].Message
	return &ChatResult{Content: m.Content, ToolCalls: m.ToolCalls, Usage: resp.Usage}, nil
}

// Embed generates embeddings for the inputs using the configured model.
func (c *HTTPClient) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	body := map[string]any{"model": c.cfg.EmbeddingModel, "input": inputs}
	var resp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := c.post(ctx, "/embeddings", body, &resp); err != nil {
		return nil, err
	}
	out := make([][]float32, len(resp.Data))
	for i, d := range resp.Data {
		out[i] = d.Embedding
	}
	return out, nil
}

// Transcribe sends audio bytes to Whisper and returns the transcript text.
func (c *HTTPClient) Transcribe(ctx context.Context, audio []byte, filename string) (string, error) {
	var resp struct {
		Text string `json:"text"`
	}
	err := c.send(ctx, func() (*http.Request, error) {
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		part, err := w.CreateFormFile("file", filename)
		if err != nil {
			return nil, err
		}
		if _, err := part.Write(audio); err != nil {
			return nil, err
		}
		_ = w.WriteField("model", c.cfg.WhisperModel)
		if err := w.Close(); err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.BaseURL+"/audio/transcriptions", &buf)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
		req.Header.Set("Content-Type", w.FormDataContentType())
		return req, nil
	}, &resp)
	if err != nil {
		return "", err
	}
	return resp.Text, nil
}
