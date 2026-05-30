package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
)

// Repository is the persistence contract consumed by the knowledge service.
type Repository interface {
	CreateKB(ctx context.Context, kb *models.KnowledgeBase) error
	GetKB(ctx context.Context, id uuid.UUID) (*models.KnowledgeBase, error)
	ListKB(ctx context.Context) ([]models.KnowledgeBase, error)
	DeleteKB(ctx context.Context, id uuid.UUID) error

	CreateDocument(ctx context.Context, d *models.Document) error
	UpdateDocument(ctx context.Context, d *models.Document) error
	GetDocument(ctx context.Context, id uuid.UUID) (*models.Document, error)
	ListDocuments(ctx context.Context, kbID uuid.UUID) ([]models.Document, error)
	DeleteDocument(ctx context.Context, id uuid.UUID) error

	ReplaceChunks(ctx context.Context, documentID uuid.UUID, chunks []models.DocumentChunk) error
	Search(ctx context.Context, kbIDs []uuid.UUID, embedding database.Vector, k int) ([]repository.ChunkResult, error)

	LinkAgentKB(ctx context.Context, agentID, kbID uuid.UUID) error
	UnlinkAgentKB(ctx context.Context, agentID, kbID uuid.UUID) error
	KBIDsForAgent(ctx context.Context, agentID uuid.UUID) ([]uuid.UUID, error)
}

// Embedder generates embeddings (satisfied by platform/openai.Client).
type Embedder interface {
	Embed(ctx context.Context, inputs []string) ([][]float32, error)
}

// Storage persists raw uploads (satisfied by platform/storage.Storage).
type Storage interface {
	Put(ctx context.Context, companyID uuid.UUID, key string, data []byte) (string, error)
	Get(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, path string) error
}
