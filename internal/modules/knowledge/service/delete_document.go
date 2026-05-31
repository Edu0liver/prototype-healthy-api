package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	"github.com/google/uuid"
)

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
