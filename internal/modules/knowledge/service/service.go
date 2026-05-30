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
	log   *zap.Logger
}

// New builds the knowledge service.
func New(repo Repository, db *database.DB, store Storage, embed Embedder, log *zap.Logger) *Service {
	return &Service{repo: repo, db: db, store: store, embed: embed, log: log}
}

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
