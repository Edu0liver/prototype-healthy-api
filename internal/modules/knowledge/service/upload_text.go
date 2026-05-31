package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/knowledge/infra/models"
	"github.com/google/uuid"
)

// UploadText stores pasted text and kicks off async indexing.
func (s *Service) UploadText(ctx context.Context, kbID uuid.UUID, title, content string) (*models.Document, error) {
	if title == "" {
		title = "text.txt"
	}
	return s.createAndIngest(ctx, kbID, "text", title, []byte(content))
}
