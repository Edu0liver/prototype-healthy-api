package service

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/google/uuid"
)

// CreateKB creates a knowledge base (RF-RAG-01).
func (s *Service) CreateKB(ctx context.Context, in dto.CreateKBRequest) (*models.KnowledgeBase, error) {
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

// GetKB loads a knowledge base by id.
func (s *Service) GetKB(ctx context.Context, id uuid.UUID) (*models.KnowledgeBase, error) {
	kb, err := s.repo.GetKB(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrKBNotFound
	}
	return kb, err
}

// ListKB returns all knowledge bases in the tenant.
func (s *Service) ListKB(ctx context.Context) ([]models.KnowledgeBase, error) {
	return s.repo.ListKB(ctx)
}

// DeleteKB removes a knowledge base.
func (s *Service) DeleteKB(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.DeleteKB(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrKBNotFound
		}
		return err
	}
	return nil
}
