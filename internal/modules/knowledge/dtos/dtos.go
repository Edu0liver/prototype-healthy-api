// Package dtos holds request/response payloads for the knowledge module.
package dtos

import "time"

// CreateKBRequest creates a knowledge base.
type CreateKBRequest struct {
	Name           string `json:"name" binding:"required,min=2"`
	Description    string `json:"description"`
	EmbeddingModel string `json:"embedding_model"`
	ChunkSize      int    `json:"chunk_size" binding:"omitempty,gt=0"`
	ChunkOverlap   int    `json:"chunk_overlap" binding:"omitempty,gte=0"`
}

// KBResponse describes a knowledge base.
type KBResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	EmbeddingModel string `json:"embedding_model"`
	ChunkSize      int    `json:"chunk_size"`
	ChunkOverlap   int    `json:"chunk_overlap"`
}

// UploadTextRequest indexes pasted text.
type UploadTextRequest struct {
	Title   string `json:"title"`
	Content string `json:"content" binding:"required"`
}

// DocumentResponse describes a document and its indexing status.
type DocumentResponse struct {
	ID         string    `json:"id"`
	Filename   string    `json:"filename"`
	SourceType string    `json:"source_type"`
	Status     string    `json:"status"`
	Error      string    `json:"error,omitempty"`
	TokenCount int       `json:"token_count"`
	CreatedAt  time.Time `json:"created_at"`
}
