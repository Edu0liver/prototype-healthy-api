package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// createAndIngest persists a document, stores its raw bytes and kicks off async
// indexing (PRD §2.7 job). Shared by UploadFile and UploadText.
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

	// Commit in its own transaction so the ingest goroutine can read the row
	// immediately — the caller's HTTP transaction may not have committed yet.
	if err := s.db.Tenant(context.Background(), companyID, func(bgCtx context.Context) error {
		return s.repo.CreateDocument(bgCtx, doc)
	}); err != nil {
		return nil, err
	}

	// Index asynchronously. Detached context + tenant scope.
	go s.ingest(companyID, doc.ID, kbID)
	return doc, nil
}

// ingest runs the extract→chunk→embed pipeline for one document (RF-RAG-03).
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
