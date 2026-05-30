package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
)

// Retrieve embeds the query and returns the top-K chunks across the agent's
// knowledge bases, always pre-filtered by company (PRD invariant 6, RF-RAG-04).
func (s *Service) Retrieve(ctx context.Context, agentID uuid.UUID, query string, k int) ([]repository.ChunkResult, error) {
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
