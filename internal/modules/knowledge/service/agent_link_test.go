package service

import (
	"context"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestLinkAgent_OK(t *testing.T) {
	linked := false
	repo := &mockRepo{linkFn: func(context.Context, uuid.UUID, uuid.UUID) error { linked = true; return nil }}
	svc := newSvc(repo, &mockStorage{}, &mockEmbedder{})
	require.NoError(t, svc.LinkAgent(context.Background(), uuid.New(), uuid.New()))
	require.True(t, linked)
}

func TestLinkAgent_KBNotFound(t *testing.T) {
	repo := &mockRepo{getKBFn: func(context.Context, uuid.UUID) (*models.KnowledgeBase, error) {
		return nil, repository.ErrNotFound
	}}
	svc := newSvc(repo, &mockStorage{}, &mockEmbedder{})
	require.ErrorIs(t, svc.LinkAgent(context.Background(), uuid.New(), uuid.New()), ErrKBNotFound)
}

func TestUnlinkAgent_OK(t *testing.T) {
	unlinked := false
	repo := &mockRepo{unlinkFn: func(context.Context, uuid.UUID, uuid.UUID) error { unlinked = true; return nil }}
	svc := newSvc(repo, &mockStorage{}, &mockEmbedder{})
	require.NoError(t, svc.UnlinkAgent(context.Background(), uuid.New(), uuid.New()))
	require.True(t, unlinked)
}

func TestKBsForAgent_OK(t *testing.T) {
	repo := &mockRepo{kbsForAgentFn: func(context.Context, uuid.UUID) ([]models.KnowledgeBase, error) {
		return []models.KnowledgeBase{{ID: uuid.New()}}, nil
	}}
	svc := newSvc(repo, &mockStorage{}, &mockEmbedder{})
	out, err := svc.KBsForAgent(context.Background(), uuid.New())
	require.NoError(t, err)
	require.Len(t, out, 1)
}
