package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

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
