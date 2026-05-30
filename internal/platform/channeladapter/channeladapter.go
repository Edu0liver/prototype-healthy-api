// Package channeladapter abstracts message channels behind one contract so the
// orchestrator, RAG and handover work identically for WhatsApp and Instagram
// (PRD §6.9). WhatsApp is backed by Evolution API; Instagram is a stub pending
// the Meta integration decision.
package channeladapter

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/platform/evolution"
)

// Channel types.
const (
	WhatsApp  = "whatsapp"
	Instagram = "instagram"
)

// Presence values.
const (
	PresenceComposing = "composing"
	PresencePaused    = "paused"
)

// ErrNotImplemented is returned by stub adapters.
var ErrNotImplemented = errors.New("channeladapter: not implemented")

// Outbound carries the per-message routing data (instance/account + creds).
type Outbound struct {
	Instance string
	APIKey   string
	Number   string // recipient (E.164 or remote id)
}

// Adapter is the uniform channel contract.
type Adapter interface {
	Channel() string
	SendText(ctx context.Context, o Outbound, text string, delayMs int) (messageID string, err error)
	SendPresence(ctx context.Context, o Outbound, presence string) error
	MarkRead(ctx context.Context, o Outbound, messageID string) error
	ConnectionState(ctx context.Context, instance string) (string, error)
}

// Registry resolves an adapter by channel type.
type Registry struct {
	adapters map[string]Adapter
}

// NewRegistry builds the registry with the WhatsApp and Instagram adapters.
func NewRegistry(evo evolution.Client) *Registry {
	r := &Registry{adapters: map[string]Adapter{}}
	r.adapters[WhatsApp] = &WhatsAppAdapter{evo: evo}
	r.adapters[Instagram] = &InstagramAdapter{}
	return r
}

// For returns the adapter for the channel type, or false if unknown.
func (r *Registry) For(channelType string) (Adapter, bool) {
	a, ok := r.adapters[channelType]
	return a, ok
}

// WhatsAppAdapter implements Adapter over Evolution API V2.
type WhatsAppAdapter struct {
	evo evolution.Client
}

func (a *WhatsAppAdapter) Channel() string { return WhatsApp }

func (a *WhatsAppAdapter) SendText(ctx context.Context, o Outbound, text string, delayMs int) (string, error) {
	res, err := a.evo.SendText(ctx, o.Instance, o.APIKey, evolution.SendTextRequest{
		Number:      o.Number,
		Text:        text,
		Delay:       delayMs,
		LinkPreview: true,
	})
	if err != nil {
		return "", err
	}
	return res.ID, nil
}

func (a *WhatsAppAdapter) SendPresence(ctx context.Context, o Outbound, presence string) error {
	return a.evo.SendPresence(ctx, o.Instance, o.APIKey, o.Number, presence)
}

func (a *WhatsAppAdapter) MarkRead(ctx context.Context, o Outbound, messageID string) error {
	return a.evo.MarkAsRead(ctx, o.Instance, o.APIKey, o.Number, messageID)
}

func (a *WhatsAppAdapter) ConnectionState(ctx context.Context, instance string) (string, error) {
	return a.evo.ConnectionState(ctx, instance)
}

// InstagramAdapter is a stub; the Meta integration is an open question (PRD §6.9).
type InstagramAdapter struct{}

func (a *InstagramAdapter) Channel() string { return Instagram }
func (a *InstagramAdapter) SendText(context.Context, Outbound, string, int) (string, error) {
	return "", ErrNotImplemented
}
func (a *InstagramAdapter) SendPresence(context.Context, Outbound, string) error {
	return ErrNotImplemented
}
func (a *InstagramAdapter) MarkRead(context.Context, Outbound, string) error {
	return ErrNotImplemented
}
func (a *InstagramAdapter) ConnectionState(context.Context, string) (string, error) {
	return "", ErrNotImplemented
}
