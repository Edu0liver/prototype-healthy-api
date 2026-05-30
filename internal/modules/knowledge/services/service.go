// Package services holds the knowledge (RAG) use cases: KB CRUD, document
// upload, asynchronous ingestion (extract→chunk→embed) and vector retrieval.
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/dtos"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repositories"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
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

// ---- Knowledge bases ------------------------------------------------------

func (s *Service) CreateKB(ctx context.Context, in dtos.CreateKBRequest) (*models.KnowledgeBase, error) {
	kb := &models.KnowledgeBase{
		ID:             uuidV7(),
		CompanyID:      appctx.CompanyID(ctx),
		Name:           in.Name,
		Description:    in.Description,
		EmbeddingModel: orDefault(in.EmbeddingModel, "text-embedding-3-small"),
		ChunkSize:      orDefaultInt(in.ChunkSize, 800),
		ChunkOverlap:   orDefaultInt(in.ChunkOverlap, 100),
	}
	if err := s.repo.CreateKB(ctx, kb); err != nil {
		return nil, err
	}
	return kb, nil
}

func (s *Service) GetKB(ctx context.Context, id uuid.UUID) (*models.KnowledgeBase, error) {
	kb, err := s.repo.GetKB(ctx, id)
	if errors.Is(err, repositories.ErrNotFound) {
		return nil, ErrKBNotFound
	}
	return kb, err
}

func (s *Service) ListKB(ctx context.Context) ([]models.KnowledgeBase, error) {
	return s.repo.ListKB(ctx)
}

func (s *Service) DeleteKB(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteKB(ctx, id); err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return ErrKBNotFound
		}
		return err
	}
	return nil
}

// ---- Documents ------------------------------------------------------------

// UploadFile stores an uploaded file and kicks off async indexing.
func (s *Service) UploadFile(ctx context.Context, kbID uuid.UUID, filename string, data []byte) (*models.Document, error) {
	return s.createAndIngest(ctx, kbID, "file", filename, data)
}

// UploadText stores pasted text and kicks off async indexing.
func (s *Service) UploadText(ctx context.Context, kbID uuid.UUID, title, content string) (*models.Document, error) {
	if title == "" {
		title = "text.txt"
	}
	return s.createAndIngest(ctx, kbID, "text", title, []byte(content))
}

func (s *Service) createAndIngest(ctx context.Context, kbID uuid.UUID, sourceType, filename string, data []byte) (*models.Document, error) {
	companyID := appctx.CompanyID(ctx)
	if _, err := s.GetKB(ctx, kbID); err != nil {
		return nil, err
	}
	doc := &models.Document{
		ID:              uuidV7(),
		CompanyID:       companyID,
		KnowledgeBaseID: kbID,
		SourceType:      sourceType,
		Filename:        filename,
		Status:          StatusPending,
	}
	path, err := s.store.Put(ctx, companyID, fmt.Sprintf("kb/%s/%s-%s", kbID, doc.ID, filename), data)
	if err != nil {
		return nil, err
	}
	doc.StoragePath = path
	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, err
	}

	// Index asynchronously (PRD §2.7 job). Detached context + tenant scope.
	go s.ingest(companyID, doc.ID, kbID)
	return doc, nil
}

func (s *Service) ListDocuments(ctx context.Context, kbID uuid.UUID) ([]models.Document, error) {
	return s.repo.ListDocuments(ctx, kbID)
}

// DeleteDocument removes a document; its chunks cascade via FK.
func (s *Service) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	doc, err := s.repo.GetDocument(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return ErrDocumentNotFound
		}
		return err
	}
	if doc.StoragePath != "" {
		_ = s.store.Delete(ctx, doc.StoragePath)
	}
	return s.repo.DeleteDocument(ctx, id)
}

// ingest runs the extract→chunk→embed pipeline for one document.
func (s *Service) ingest(companyID, documentID, kbID uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err := s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
		doc, err := s.repo.GetDocument(ctx, documentID)
		if err != nil {
			return err
		}
		kb, err := s.repo.GetKB(ctx, kbID)
		if err != nil {
			return err
		}
		doc.Status = StatusProcessing
		if err := s.repo.UpdateDocument(ctx, doc); err != nil {
			return err
		}

		raw, err := s.store.Get(ctx, doc.StoragePath)
		if err != nil {
			return err
		}
		text, err := extractText(doc.Filename, raw)
		if err != nil {
			return err
		}
		pieces := chunkText(text, kb.ChunkSize, kb.ChunkOverlap)
		if len(pieces) == 0 {
			doc.Status = StatusIndexed
			doc.TokenCount = 0
			return s.repo.UpdateDocument(ctx, doc)
		}

		vectors, err := s.embed.Embed(ctx, pieces)
		if err != nil {
			return err
		}
		if len(vectors) != len(pieces) {
			return fmt.Errorf("knowledge: embedding count mismatch (%d != %d)", len(vectors), len(pieces))
		}

		chunks := make([]models.DocumentChunk, len(pieces))
		total := 0
		for i, p := range pieces {
			chunks[i] = models.DocumentChunk{
				ID:              uuidV7(),
				CompanyID:       companyID,
				KnowledgeBaseID: kbID,
				DocumentID:      documentID,
				ChunkIndex:      i,
				Content:         p,
				Embedding:       database.Vector(vectors[i]),
				Metadata:        map[string]any{"filename": doc.Filename},
			}
			total += estimateTokens(p)
		}
		if err := s.repo.ReplaceChunks(ctx, documentID, chunks); err != nil {
			return err
		}
		doc.Status = StatusIndexed
		doc.TokenCount = total
		return s.repo.UpdateDocument(ctx, doc)
	})

	if err != nil {
		s.log.Warn("rag ingestion failed", zap.String("document_id", documentID.String()), zap.Error(err))
		s.markFailed(companyID, documentID, err)
	}
}

func (s *Service) markFailed(companyID, documentID uuid.UUID, cause error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
		doc, err := s.repo.GetDocument(ctx, documentID)
		if err != nil {
			return err
		}
		doc.Status = StatusFailed
		doc.Error = cause.Error()
		return s.repo.UpdateDocument(ctx, doc)
	})
}

// ---- Retrieval (used by orchestration) ------------------------------------

// Retrieve embeds the query and returns the top-K chunks across the agent's
// knowledge bases, always pre-filtered by company (PRD invariant 6, RF-RAG-04).
func (s *Service) Retrieve(ctx context.Context, agentID uuid.UUID, query string, k int) ([]repositories.ChunkResult, error) {
	kbIDs, err := s.repo.KBIDsForAgent(ctx, agentID)
	if err != nil || len(kbIDs) == 0 {
		return nil, err
	}
	vectors, err := s.embed.Embed(ctx, []string{query})
	if err != nil || len(vectors) == 0 {
		return nil, err
	}
	return s.repo.Search(ctx, kbIDs, database.Vector(vectors[0]), k)
}

// ---- Agent ↔ KB -----------------------------------------------------------

func (s *Service) LinkAgent(ctx context.Context, agentID, kbID uuid.UUID) error {
	if _, err := s.GetKB(ctx, kbID); err != nil {
		return err
	}
	return s.repo.LinkAgentKB(ctx, agentID, kbID)
}

func (s *Service) UnlinkAgent(ctx context.Context, agentID, kbID uuid.UUID) error {
	return s.repo.UnlinkAgentKB(ctx, agentID, kbID)
}

func (s *Service) KBIDsForAgent(ctx context.Context, agentID uuid.UUID) ([]uuid.UUID, error) {
	return s.repo.KBIDsForAgent(ctx, agentID)
}

// ---- helpers --------------------------------------------------------------

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
