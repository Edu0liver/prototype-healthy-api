// Package service holds the tenant use cases (one file per use case).
package service

import (
	"strings"

	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/google/uuid"
)

// Service implements tenant + white-label use cases. Public operations
// (signup, branding-by-host) run through db.System (no tenant context yet);
// authenticated operations use the request's tenant transaction in context.
type Service struct {
	repo Repository
	db   *database.DB
}

// New builds the tenant service.
func New(repo Repository, db *database.DB) *Service {
	return &Service{repo: repo, db: db}
}

func mustUUIDv7() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate key value")
}
