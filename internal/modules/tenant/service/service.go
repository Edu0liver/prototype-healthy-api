// Package services holds the tenant use cases.
package service

import "github.com/Edu0liver/prototype-healthy-api/internal/shared/database"

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
