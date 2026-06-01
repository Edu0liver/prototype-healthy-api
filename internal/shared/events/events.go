// Package events publishes realtime domain events to Redis Pub/Sub, which the
// realtime module bridges to WebSocket clients. This keeps producers
// (conversation, handover, orchestration) decoupled from the WebSocket hub.
package events

import (
	"context"
	"encoding/json"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Event types.
const (
	TypeMessage  = "message"
	TypeState    = "state"
	TypeQRUpdate = "qr_update"
)

// Event is a realtime notification scoped to a company.
type Event struct {
	Type           string         `json:"type"`
	ConversationID string         `json:"conversation_id"`
	Payload        map[string]any `json:"payload,omitempty"`
}

// Publisher publishes events to per-company Redis channels.
type Publisher struct {
	rdb *redisx.Client
	log *zap.Logger
}

// New builds the publisher.
func New(rdb *redisx.Client, log *zap.Logger) *Publisher {
	return &Publisher{rdb: rdb, log: log}
}

// Channel is the Redis Pub/Sub channel for a company's realtime stream.
func Channel(companyID uuid.UUID) string { return "rt:" + companyID.String() }

// Publish sends an event to the company's realtime channel (best effort).
func (p *Publisher) Publish(ctx context.Context, companyID uuid.UUID, ev Event) {
	data, err := json.Marshal(ev)
	if err != nil {
		return
	}
	if err := p.rdb.Publish(ctx, Channel(companyID), data).Err(); err != nil {
		p.log.Debug("events publish failed", zap.Error(err))
	}
}
