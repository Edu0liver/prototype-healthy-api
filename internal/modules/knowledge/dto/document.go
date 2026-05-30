package dto

import "time"

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
