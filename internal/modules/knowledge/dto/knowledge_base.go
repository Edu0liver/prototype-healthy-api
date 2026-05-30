// Package dto holds request/response payloads for the knowledge module.
package dto

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
