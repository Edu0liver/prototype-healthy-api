// Package evolution is a thin client over Evolution API V2 (WhatsApp gateway).
// Reference: https://doc.evolution-api.com/v2/api-reference
package evolution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/pkg/resilience"
)

// CreateInstanceRequest is the payload for POST /instance/create.
type CreateInstanceRequest struct {
	InstanceName string         `json:"instanceName"`
	Integration  string         `json:"integration"` // WHATSAPP-BAILEYS | WHATSAPP-BUSINESS
	QRCode       bool           `json:"qrcode"`
	Number       string         `json:"number,omitempty"`
	Webhook      *WebhookConfig `json:"webhook,omitempty"`
}

// WebhookConfig configures Evolution to post events to our ingestion endpoint.
type WebhookConfig struct {
	URL      string            `json:"url"`
	ByEvents bool              `json:"byEvents"`
	Base64   bool              `json:"base64"`
	Headers  map[string]string `json:"headers,omitempty"`
	Events   []string          `json:"events"`
}

// CreateInstanceResult is the normalized create response.
type CreateInstanceResult struct {
	InstanceName string
	InstanceID   string
	Status       string
	APIKey       string
}

// ConnectResult holds QR / pairing connection data.
type ConnectResult struct {
	Code        string `json:"code"`        // QR content
	PairingCode string `json:"pairingCode"` // manual pairing code
	Count       int    `json:"count"`
}

// SendTextRequest is the payload for POST /message/sendText/{instance}.
type SendTextRequest struct {
	Number      string `json:"number"`
	Text        string `json:"text"`
	Delay       int    `json:"delay,omitempty"`
	LinkPreview bool   `json:"linkPreview"`
}

// SendResult is the normalized send response.
type SendResult struct {
	ID     string
	Status string
}

// Client is the Evolution API contract (interface for mocking).
type Client interface {
	CreateInstance(ctx context.Context, req CreateInstanceRequest) (*CreateInstanceResult, error)
	Connect(ctx context.Context, instance, number string) (*ConnectResult, error)
	ConnectionState(ctx context.Context, instance string) (string, error)
	Logout(ctx context.Context, instance string) error
	DeleteInstance(ctx context.Context, instance string) error
	SendText(ctx context.Context, instance, apiKey string, req SendTextRequest) (*SendResult, error)
	SendPresence(ctx context.Context, instance, apiKey, number, presence string) error
	MarkAsRead(ctx context.Context, instance, apiKey, remoteJID, messageID string) error
	GetMediaBase64(ctx context.Context, instance, apiKey, messageID string) (data string, mimetype string, err error)
}

// Config configures the Evolution client.
type Config struct {
	BaseURL      string
	GlobalAPIKey string
	Timeout      time.Duration
}

// HTTPClient is the live implementation.
type HTTPClient struct {
	cfg  Config
	http *http.Client
}

// New builds the live Evolution client.
func New(cfg Config) *HTTPClient {
	return &HTTPClient{cfg: cfg, http: &http.Client{Timeout: cfg.Timeout}}
}

func (c *HTTPClient) request(ctx context.Context, method, path, apiKey string, body, out any) error {
	var payload []byte
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = buf
	}
	key := apiKey
	if key == "" {
		key = c.cfg.GlobalAPIKey
	}
	// Retry network errors and 429/5xx with backoff (NFR §5.4); 4xx are permanent.
	return resilience.Do(ctx, resilience.DefaultRetry(), func() error {
		var reader io.Reader
		if payload != nil {
			reader = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, c.cfg.BaseURL+path, reader)
		if err != nil {
			return resilience.Permanent(err)
		}
		req.Header.Set("apikey", key)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		data, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			return fmt.Errorf("evolution: %s %s: %s: %s", method, path, resp.Status, string(data))
		}
		if resp.StatusCode >= 300 {
			return resilience.Permanent(fmt.Errorf("evolution: %s %s: %s: %s", method, path, resp.Status, string(data)))
		}
		if out != nil {
			return json.Unmarshal(data, out)
		}
		return nil
	})
}

// CreateInstance provisions a WhatsApp instance.
func (c *HTTPClient) CreateInstance(ctx context.Context, req CreateInstanceRequest) (*CreateInstanceResult, error) {
	var resp struct {
		Instance struct {
			InstanceName string `json:"instanceName"`
			InstanceID   string `json:"instanceId"`
			Status       string `json:"status"`
		} `json:"instance"`
		Hash struct {
			APIKey string `json:"apikey"`
		} `json:"hash"`
	}
	// Evolution has historically returned `hash` as either an object or a string;
	// decode into a generic map first for robustness.
	var raw map[string]json.RawMessage
	if err := c.request(ctx, http.MethodPost, "/instance/create", "", req, &raw); err != nil {
		return nil, err
	}
	if v, ok := raw["instance"]; ok {
		_ = json.Unmarshal(v, &resp.Instance)
	}
	if v, ok := raw["hash"]; ok {
		if err := json.Unmarshal(v, &resp.Hash); err != nil {
			// hash returned as bare string
			var s string
			if json.Unmarshal(v, &s) == nil {
				resp.Hash.APIKey = s
			}
		}
	}
	return &CreateInstanceResult{
		InstanceName: resp.Instance.InstanceName,
		InstanceID:   resp.Instance.InstanceID,
		Status:       resp.Instance.Status,
		APIKey:       resp.Hash.APIKey,
	}, nil
}

// Connect returns QR (no number) or pairing code (with number).
func (c *HTTPClient) Connect(ctx context.Context, instance, number string) (*ConnectResult, error) {
	path := "/instance/connect/" + instance
	if number != "" {
		path += "?number=" + number
	}
	var res ConnectResult
	if err := c.request(ctx, http.MethodGet, path, "", nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

// ConnectionState returns the instance state (open|connecting|close).
func (c *HTTPClient) ConnectionState(ctx context.Context, instance string) (string, error) {
	var res struct {
		Instance struct {
			State string `json:"state"`
		} `json:"instance"`
	}
	if err := c.request(ctx, http.MethodGet, "/instance/connectionState/"+instance, "", nil, &res); err != nil {
		return "", err
	}
	return res.Instance.State, nil
}

// Logout ends the WhatsApp session.
func (c *HTTPClient) Logout(ctx context.Context, instance string) error {
	return c.request(ctx, http.MethodDelete, "/instance/logout/"+instance, "", nil, nil)
}

// DeleteInstance removes the instance.
func (c *HTTPClient) DeleteInstance(ctx context.Context, instance string) error {
	return c.request(ctx, http.MethodDelete, "/instance/delete/"+instance, "", nil, nil)
}

// SendText sends a text message.
func (c *HTTPClient) SendText(ctx context.Context, instance, apiKey string, req SendTextRequest) (*SendResult, error) {
	var res struct {
		Key struct {
			ID string `json:"id"`
		} `json:"key"`
		Status string `json:"status"`
	}
	if err := c.request(ctx, http.MethodPost, "/message/sendText/"+instance, apiKey, req, &res); err != nil {
		return nil, err
	}
	return &SendResult{ID: res.Key.ID, Status: res.Status}, nil
}

// SendPresence emits a presence update (e.g. "composing").
func (c *HTTPClient) SendPresence(ctx context.Context, instance, apiKey, number, presence string) error {
	body := map[string]any{"number": number, "presence": presence}
	return c.request(ctx, http.MethodPost, "/chat/sendPresence/"+instance, apiKey, body, nil)
}

// MarkAsRead marks a message as read.
func (c *HTTPClient) MarkAsRead(ctx context.Context, instance, apiKey, remoteJID, messageID string) error {
	body := map[string]any{
		"readMessages": []map[string]any{
			{"remoteJid": remoteJID, "fromMe": false, "id": messageID},
		},
	}
	return c.request(ctx, http.MethodPost, "/chat/markMessageAsRead/"+instance, apiKey, body, nil)
}

// GetMediaBase64 fetches media bytes (base64) for a received message.
func (c *HTTPClient) GetMediaBase64(ctx context.Context, instance, apiKey, messageID string) (string, string, error) {
	body := map[string]any{"message": map[string]any{"key": map[string]any{"id": messageID}}}
	var res struct {
		Base64   string `json:"base64"`
		Mimetype string `json:"mimetype"`
	}
	if err := c.request(ctx, http.MethodPost, "/chat/getBase64FromMediaMessage/"+instance, apiKey, body, &res); err != nil {
		return "", "", err
	}
	return res.Base64, res.Mimetype, nil
}
