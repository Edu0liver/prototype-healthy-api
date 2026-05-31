package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// mockRepo is a function-backed knowledge Repository. Only the methods a test
// needs are wired; the rest return zero values.
type mockRepo struct {
	createKBFn      func(ctx context.Context, kb *models.KnowledgeBase) error
	getKBFn         func(ctx context.Context, id uuid.UUID) (*models.KnowledgeBase, error)
	listKBFn        func(ctx context.Context) ([]models.KnowledgeBase, error)
	deleteKBFn      func(ctx context.Context, id uuid.UUID) error
	createDocFn     func(ctx context.Context, d *models.Document) error
	updateDocFn     func(ctx context.Context, d *models.Document) error
	getDocFn        func(ctx context.Context, id uuid.UUID) (*models.Document, error)
	listDocsFn      func(ctx context.Context, kbID uuid.UUID) ([]models.Document, error)
	deleteDocFn     func(ctx context.Context, id uuid.UUID) error
	replaceChunksFn func(ctx context.Context, documentID uuid.UUID, chunks []models.DocumentChunk) error
	searchFn        func(ctx context.Context, kbIDs []uuid.UUID, embedding database.Vector, k int) ([]repository.ChunkResult, error)
	linkFn          func(ctx context.Context, agentID, kbID uuid.UUID) error
	unlinkFn        func(ctx context.Context, agentID, kbID uuid.UUID) error
	kbIDsFn         func(ctx context.Context, agentID uuid.UUID) ([]uuid.UUID, error)
	kbsForAgentFn   func(ctx context.Context, agentID uuid.UUID) ([]models.KnowledgeBase, error)
}

func (m *mockRepo) CreateKB(ctx context.Context, kb *models.KnowledgeBase) error {
	if m.createKBFn != nil {
		return m.createKBFn(ctx, kb)
	}
	return nil
}
func (m *mockRepo) GetKB(ctx context.Context, id uuid.UUID) (*models.KnowledgeBase, error) {
	if m.getKBFn != nil {
		return m.getKBFn(ctx, id)
	}
	return &models.KnowledgeBase{ID: id}, nil
}
func (m *mockRepo) ListKB(ctx context.Context) ([]models.KnowledgeBase, error) {
	if m.listKBFn != nil {
		return m.listKBFn(ctx)
	}
	return nil, nil
}
func (m *mockRepo) DeleteKB(ctx context.Context, id uuid.UUID) error {
	if m.deleteKBFn != nil {
		return m.deleteKBFn(ctx, id)
	}
	return nil
}
func (m *mockRepo) CreateDocument(ctx context.Context, d *models.Document) error {
	if m.createDocFn != nil {
		return m.createDocFn(ctx, d)
	}
	return nil
}
func (m *mockRepo) UpdateDocument(ctx context.Context, d *models.Document) error {
	if m.updateDocFn != nil {
		return m.updateDocFn(ctx, d)
	}
	return nil
}
func (m *mockRepo) GetDocument(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	if m.getDocFn != nil {
		return m.getDocFn(ctx, id)
	}
	return &models.Document{ID: id}, nil
}
func (m *mockRepo) ListDocuments(ctx context.Context, kbID uuid.UUID) ([]models.Document, error) {
	if m.listDocsFn != nil {
		return m.listDocsFn(ctx, kbID)
	}
	return nil, nil
}
func (m *mockRepo) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	if m.deleteDocFn != nil {
		return m.deleteDocFn(ctx, id)
	}
	return nil
}
func (m *mockRepo) ReplaceChunks(ctx context.Context, documentID uuid.UUID, chunks []models.DocumentChunk) error {
	if m.replaceChunksFn != nil {
		return m.replaceChunksFn(ctx, documentID, chunks)
	}
	return nil
}
func (m *mockRepo) Search(ctx context.Context, kbIDs []uuid.UUID, embedding database.Vector, k int) ([]repository.ChunkResult, error) {
	if m.searchFn != nil {
		return m.searchFn(ctx, kbIDs, embedding, k)
	}
	return nil, nil
}
func (m *mockRepo) LinkAgentKB(ctx context.Context, agentID, kbID uuid.UUID) error {
	if m.linkFn != nil {
		return m.linkFn(ctx, agentID, kbID)
	}
	return nil
}
func (m *mockRepo) UnlinkAgentKB(ctx context.Context, agentID, kbID uuid.UUID) error {
	if m.unlinkFn != nil {
		return m.unlinkFn(ctx, agentID, kbID)
	}
	return nil
}
func (m *mockRepo) KBIDsForAgent(ctx context.Context, agentID uuid.UUID) ([]uuid.UUID, error) {
	if m.kbIDsFn != nil {
		return m.kbIDsFn(ctx, agentID)
	}
	return nil, nil
}
func (m *mockRepo) KBsForAgent(ctx context.Context, agentID uuid.UUID) ([]models.KnowledgeBase, error) {
	if m.kbsForAgentFn != nil {
		return m.kbsForAgentFn(ctx, agentID)
	}
	return nil, nil
}

// mockEmbedder is a function-backed Embedder.
type mockEmbedder struct {
	embedFn func(ctx context.Context, inputs []string) ([][]float32, error)
}

func (m *mockEmbedder) Embed(ctx context.Context, inputs []string) ([][]float32, error) {
	if m.embedFn != nil {
		return m.embedFn(ctx, inputs)
	}
	return make([][]float32, len(inputs)), nil
}

// mockStorage is a function-backed Storage.
type mockStorage struct {
	putFn    func(ctx context.Context, companyID uuid.UUID, key string, data []byte) (string, error)
	getFn    func(ctx context.Context, path string) ([]byte, error)
	deleteFn func(ctx context.Context, path string) error
}

func (m *mockStorage) Put(ctx context.Context, companyID uuid.UUID, key string, data []byte) (string, error) {
	if m.putFn != nil {
		return m.putFn(ctx, companyID, key, data)
	}
	return key, nil
}
func (m *mockStorage) Get(ctx context.Context, path string) ([]byte, error) {
	if m.getFn != nil {
		return m.getFn(ctx, path)
	}
	return nil, nil
}
func (m *mockStorage) Delete(ctx context.Context, path string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, path)
	}
	return nil
}

// newSvc builds a knowledge service with a nil DB (use cases under test must not
// touch db.System/db.Tenant) and a nop logger.
func newSvc(repo Repository, store Storage, embed Embedder) *Service {
	return New(repo, nil, store, embed, zap.NewNop())
}
