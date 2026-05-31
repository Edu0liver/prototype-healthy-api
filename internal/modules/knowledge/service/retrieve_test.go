package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRetrieve_NoKBs_SkipsEmbed(t *testing.T) {
	embedCalled := false
	repo := &mockRepo{kbIDsFn: func(context.Context, uuid.UUID) ([]uuid.UUID, error) { return nil, nil }}
	embed := &mockEmbedder{embedFn: func(context.Context, []string) ([][]float32, error) {
		embedCalled = true
		return nil, nil
	}}
	svc := newSvc(repo, &mockStorage{}, embed)
	out, err := svc.Retrieve(context.Background(), uuid.New(), "q", 5)
	require.NoError(t, err)
	require.Nil(t, out)
	require.False(t, embedCalled, "no linked KBs => must not call the embedder")
}

func TestRetrieve_HappyPath(t *testing.T) {
	kbID := uuid.New()
	repo := &mockRepo{
		kbIDsFn: func(context.Context, uuid.UUID) ([]uuid.UUID, error) { return []uuid.UUID{kbID}, nil },
		searchFn: func(_ context.Context, kbIDs []uuid.UUID, _ database.Vector, k int) ([]repository.ChunkResult, error) {
			require.Equal(t, []uuid.UUID{kbID}, kbIDs)
			require.Equal(t, 3, k)
			return []repository.ChunkResult{{Content: "hit", Score: 0.9}}, nil
		},
	}
	embed := &mockEmbedder{embedFn: func(context.Context, []string) ([][]float32, error) {
		return [][]float32{{0.1, 0.2, 0.3}}, nil
	}}
	svc := newSvc(repo, &mockStorage{}, embed)
	out, err := svc.Retrieve(context.Background(), uuid.New(), "q", 3)
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "hit", out[0].Content)
}

func TestRetrieve_EmbedError(t *testing.T) {
	boom := errors.New("openai down")
	repo := &mockRepo{kbIDsFn: func(context.Context, uuid.UUID) ([]uuid.UUID, error) { return []uuid.UUID{uuid.New()}, nil }}
	embed := &mockEmbedder{embedFn: func(context.Context, []string) ([][]float32, error) { return nil, boom }}
	svc := newSvc(repo, &mockStorage{}, embed)
	out, err := svc.Retrieve(context.Background(), uuid.New(), "q", 5)
	require.Nil(t, out)
	require.ErrorIs(t, err, boom)
}
