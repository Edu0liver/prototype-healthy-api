package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/dto"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func kbCtx() context.Context {
	return appctx.With(context.Background(), appctx.Identity{CompanyID: uuid.New()})
}

func TestCreateKB_Defaults(t *testing.T) {
	var saved *models.KnowledgeBase
	svc := newSvc(&mockRepo{createKBFn: func(_ context.Context, kb *models.KnowledgeBase) error {
		saved = kb
		return nil
	}}, &mockStorage{}, &mockEmbedder{})

	out, err := svc.CreateKB(kbCtx(), dto.CreateKBRequest{Name: "Docs"})
	require.NoError(t, err)
	require.Equal(t, "text-embedding-3-small", out.EmbeddingModel)
	require.Equal(t, 800, out.ChunkSize)
	require.Equal(t, 100, out.ChunkOverlap)
	require.NotEqual(t, uuid.Nil, out.ID)
	require.Same(t, saved, out)
}

func TestCreateKB_RepoError(t *testing.T) {
	boom := errors.New("insert failed")
	svc := newSvc(&mockRepo{createKBFn: func(context.Context, *models.KnowledgeBase) error { return boom }}, &mockStorage{}, &mockEmbedder{})
	out, err := svc.CreateKB(kbCtx(), dto.CreateKBRequest{Name: "Docs"})
	require.Nil(t, out)
	require.ErrorIs(t, err, boom)
}

func TestGetKB_NotFoundMapped(t *testing.T) {
	svc := newSvc(&mockRepo{getKBFn: func(context.Context, uuid.UUID) (*models.KnowledgeBase, error) {
		return nil, repository.ErrNotFound
	}}, &mockStorage{}, &mockEmbedder{})
	out, err := svc.GetKB(context.Background(), uuid.New())
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrKBNotFound)
}

func TestListKB_OK(t *testing.T) {
	svc := newSvc(&mockRepo{listKBFn: func(context.Context) ([]models.KnowledgeBase, error) {
		return []models.KnowledgeBase{{ID: uuid.New()}}, nil
	}}, &mockStorage{}, &mockEmbedder{})
	out, err := svc.ListKB(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 1)
}

func TestDeleteKB_NotFoundMapped(t *testing.T) {
	svc := newSvc(&mockRepo{deleteKBFn: func(context.Context, uuid.UUID) error { return repository.ErrNotFound }}, &mockStorage{}, &mockEmbedder{})
	require.ErrorIs(t, svc.DeleteKB(context.Background(), uuid.New()), ErrKBNotFound)
}
