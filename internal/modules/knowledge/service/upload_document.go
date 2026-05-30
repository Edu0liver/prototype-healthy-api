package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/google/uuid"
)

// UploadFile stores an uploaded file and kicks off async indexing (RF-RAG-02).
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

// ListDocuments returns documents of a knowledge base.
func (s *Service) ListDocuments(ctx context.Context, kbID uuid.UUID) ([]models.Document, error) {
	return s.repo.ListDocuments(ctx, kbID)
}

// DeleteDocument removes a document; its chunks cascade via FK (RF-RAG-05).
func (s *Service) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	doc, err := s.repo.GetDocument(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrDocumentNotFound
		}
		return err
	}
	if doc.StoragePath != "" {
		_ = s.store.Delete(ctx, doc.StoragePath)
	}
	return s.repo.DeleteDocument(ctx, id)
}
