package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestListDocuments_OK(t *testing.T) {
	kbID := uuid.New()
	svc := newSvc(&mockRepo{listDocsFn: func(_ context.Context, gotKB uuid.UUID) ([]models.Document, error) {
		require.Equal(t, kbID, gotKB)
		return []models.Document{{ID: uuid.New()}, {ID: uuid.New()}}, nil
	}}, &mockStorage{}, &mockEmbedder{})
	out, err := svc.ListDocuments(context.Background(), kbID)
	require.NoError(t, err)
	require.Len(t, out, 2)
}

func TestDeleteDocument_NotFoundMapped(t *testing.T) {
	svc := newSvc(&mockRepo{getDocFn: func(context.Context, uuid.UUID) (*models.Document, error) {
		return nil, repository.ErrNotFound
	}}, &mockStorage{}, &mockEmbedder{})
	require.ErrorIs(t, svc.DeleteDocument(context.Background(), uuid.New()), ErrDocumentNotFound)
}

func TestDeleteDocument_RemovesStoredBlob(t *testing.T) {
	deletedPath := ""
	repo := &mockRepo{getDocFn: func(_ context.Context, id uuid.UUID) (*models.Document, error) {
		return &models.Document{ID: id, StoragePath: "kb/x/doc.txt"}, nil
	}}
	store := &mockStorage{deleteFn: func(_ context.Context, path string) error { deletedPath = path; return nil }}
	svc := newSvc(repo, store, &mockEmbedder{})
	require.NoError(t, svc.DeleteDocument(context.Background(), uuid.New()))
	require.Equal(t, "kb/x/doc.txt", deletedPath, "stored blob must be deleted alongside the row")
}

func TestDeleteDocument_RepoError(t *testing.T) {
	boom := errors.New("delete failed")
	repo := &mockRepo{
		getDocFn:    func(_ context.Context, id uuid.UUID) (*models.Document, error) { return &models.Document{ID: id}, nil },
		deleteDocFn: func(context.Context, uuid.UUID) error { return boom },
	}
	svc := newSvc(repo, &mockStorage{}, &mockEmbedder{})
	require.ErrorIs(t, svc.DeleteDocument(context.Background(), uuid.New()), boom)
}
