// Package models holds the GORM entities for the knowledge (RAG) module.
package models

import (
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/google/uuid"
)

// KnowledgeBase groups documents indexed under one embedding configuration.
type KnowledgeBase struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID      uuid.UUID `gorm:"type:uuid"`
	Name           string
	Description    string
	EmbeddingModel string
	ChunkSize      int
	ChunkOverlap   int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (KnowledgeBase) TableName() string { return "knowledge_bases" }

// Document is an uploaded file or pasted text awaiting / having indexing.
type Document struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID       uuid.UUID `gorm:"type:uuid"`
	KnowledgeBaseID uuid.UUID `gorm:"type:uuid"`
	SourceType      string    // file | text
	Filename        string
	StoragePath     string
	Status          string // pending | processing | indexed | failed
	Error           string
	TokenCount      int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (Document) TableName() string { return "documents" }

// DocumentChunk is an embedded text fragment for vector retrieval.
type DocumentChunk struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID       uuid.UUID `gorm:"type:uuid"`
	KnowledgeBaseID uuid.UUID `gorm:"type:uuid"`
	DocumentID      uuid.UUID `gorm:"type:uuid"`
	ChunkIndex      int
	Content         string
	Embedding       database.Vector  `gorm:"type:vector(1536)"`
	Metadata        database.JSONMap `gorm:"type:jsonb"`
	CreatedAt       time.Time
}

func (DocumentChunk) TableName() string { return "document_chunks" }

// AgentKnowledgeBase is the N:M join between agents and knowledge bases.
type AgentKnowledgeBase struct {
	AgentID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	KnowledgeBaseID uuid.UUID `gorm:"type:uuid;primaryKey"`
	CompanyID       uuid.UUID `gorm:"type:uuid"`
}

func (AgentKnowledgeBase) TableName() string { return "agent_knowledge_bases" }
