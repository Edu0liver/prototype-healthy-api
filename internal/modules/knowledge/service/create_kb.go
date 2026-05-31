package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
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
