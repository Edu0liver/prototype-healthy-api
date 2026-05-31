// Package repositories implements tenant-scoped persistence for the knowledge
// (RAG) module, including the pgvector similarity search.
package repository

import (
	"context"
	"errors"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrNotFound indicates the row does not exist in the tenant scope.
var ErrNotFound = errors.New("knowledge: not found")

// Repository persists knowledge bases, documents and chunks.
type Repository struct{}

// New builds the repository.
func New() *Repository { return &Repository{} }

func wrap(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

// ---- Knowledge bases ------------------------------------------------------

func (r *Repository) CreateKB(ctx context.Context, kb *models.KnowledgeBase) error {
	return wrap(database.MustTx(ctx).Create(kb).Error)
}

func (r *Repository) GetKB(ctx context.Context, id uuid.UUID) (*models.KnowledgeBase, error) {
	var kb models.KnowledgeBase
	if err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).First(&kb, "id = ?", id).Error; err != nil {
		return nil, wrap(err)
	}
	return &kb, nil
}

func (r *Repository) ListKB(ctx context.Context) ([]models.KnowledgeBase, error) {
	var out []models.KnowledgeBase
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Order("created_at DESC").Find(&out).Error
	return out, err
}

func (r *Repository) DeleteKB(ctx context.Context, id uuid.UUID) error {
	res := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Delete(&models.KnowledgeBase{}, "id = ?", id)
	if res.Error != nil {
		return wrap(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ---- Documents ------------------------------------------------------------

func (r *Repository) CreateDocument(ctx context.Context, d *models.Document) error {
	return wrap(database.MustTx(ctx).Create(d).Error)
}

func (r *Repository) UpdateDocument(ctx context.Context, d *models.Document) error {
	return wrap(database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Save(d).Error)
}

func (r *Repository) GetDocument(ctx context.Context, id uuid.UUID) (*models.Document, error) {
	var d models.Document
	if err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).First(&d, "id = ?", id).Error; err != nil {
		return nil, wrap(err)
	}
	return &d, nil
}

func (r *Repository) ListDocuments(ctx context.Context, kbID uuid.UUID) ([]models.Document, error) {
	var out []models.Document
	err := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).
		Where("knowledge_base_id = ?", kbID).Order("created_at DESC").Find(&out).Error
	return out, err
}

func (r *Repository) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	res := database.MustTx(ctx).Scopes(database.TenantScope(ctx)).Delete(&models.Document{}, "id = ?", id)
	if res.Error != nil {
		return wrap(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ---- Chunks ---------------------------------------------------------------

// ReplaceChunks deletes existing chunks for a document and inserts the new set.
func (r *Repository) ReplaceChunks(ctx context.Context, documentID uuid.UUID, chunks []models.DocumentChunk) error {
	tx := database.MustTx(ctx)
	if err := tx.Scopes(database.TenantScope(ctx)).Delete(&models.DocumentChunk{}, "document_id = ?", documentID).Error; err != nil {
		return err
	}
	if len(chunks) == 0 {
		return nil
	}
	return tx.Create(&chunks).Error
}

// ChunkResult is a retrieval hit.
type ChunkResult struct {
	ID         uuid.UUID
	DocumentID uuid.UUID
	Content    string
	Score      float64
}

// Search returns the top-K most similar chunks, pre-filtered by company and the
// agent's knowledge bases (PRD invariant 6).
func (r *Repository) Search(ctx context.Context, kbIDs []uuid.UUID, embedding database.Vector, k int) ([]ChunkResult, error) {
	if len(kbIDs) == 0 {
		return nil, nil
	}
	companyID := appctx.CompanyID(ctx)
	var rows []ChunkResult
	err := database.MustTx(ctx).Raw(
		`SELECT id, document_id, content, 1 - (embedding <=> ?) AS score
		   FROM document_chunks
		  WHERE company_id = ? AND knowledge_base_id IN ?
		  ORDER BY embedding <=> ?
		  LIMIT ?`,
		embedding, companyID, kbIDs, embedding, k,
	).Scan(&rows).Error
	return rows, err
}

// ---- Agent ↔ KB (N:M) -----------------------------------------------------

func (r *Repository) LinkAgentKB(ctx context.Context, agentID, kbID uuid.UUID) error {
	link := models.AgentKnowledgeBase{AgentID: agentID, KnowledgeBaseID: kbID, CompanyID: appctx.CompanyID(ctx)}
	return database.MustTx(ctx).Create(&link).Error
}

func (r *Repository) UnlinkAgentKB(ctx context.Context, agentID, kbID uuid.UUID) error {
	return database.MustTx(ctx).Scopes(database.TenantScope(ctx)).
		Delete(&models.AgentKnowledgeBase{}, "agent_id = ? AND knowledge_base_id = ?", agentID, kbID).Error
}

// KBIDsForAgent returns just the knowledge base IDs linked to an agent (used by retrieval).
func (r *Repository) KBIDsForAgent(ctx context.Context, agentID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := database.MustTx(ctx).Model(&models.AgentKnowledgeBase{}).
		Scopes(database.TenantScope(ctx)).
		Where("agent_id = ?", agentID).Pluck("knowledge_base_id", &ids).Error
	return ids, err
}

// KBsForAgent returns full knowledge base records linked to an agent.
// TenantScope cannot be used here because both tables share a company_id column,
// causing an ambiguous reference error on JOIN; qualify explicitly instead.
func (r *Repository) KBsForAgent(ctx context.Context, agentID uuid.UUID) ([]models.KnowledgeBase, error) {
	companyID := appctx.CompanyID(ctx)
	var out []models.KnowledgeBase
	err := database.MustTx(ctx).
		Where("knowledge_bases.company_id = ?", companyID).
		Joins("JOIN agent_knowledge_bases akb ON akb.knowledge_base_id = knowledge_bases.id").
		Where("akb.agent_id = ?", agentID).
		Order("knowledge_bases.created_at DESC").
		Find(&out).Error
	return out, err
}
