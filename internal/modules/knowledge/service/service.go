// Package service holds the knowledge (RAG) use cases: KB CRUD, document upload,
// asynchronous ingestion (extract→chunk→embed) and vector retrieval.
package service

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Document statuses.
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusIndexed    = "indexed"
	StatusFailed     = "failed"
)

// Service implements the knowledge use cases.
type Service struct {
	repo  Repository
	db    *database.DB
	store Storage
	embed Embedder
	bill  Billing
	log   *zap.Logger
}

// New builds the knowledge service. Billing (quota + metering) defaults to a
// no-op; the fx module wires the real one via WithBilling so unit tests need no
// billing dependency.
func New(repo Repository, db *database.DB, store Storage, embed Embedder, log *zap.Logger) *Service {
	return &Service{repo: repo, db: db, store: store, embed: embed, bill: noopBilling{}, log: log}
}

// WithBilling installs the billing quota/metering hooks (production wiring).
func (s *Service) WithBilling(b Billing) *Service { s.bill = b; return s }

func uuidV7() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}
func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
func orDefaultInt(v, def int) int {
	if v == 0 {
		return def
	}
	return v
}
