// Package services holds the iam use cases (auth, invites, user management).
package services

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/platform/token"
)

// Service implements the iam use cases.
type Service struct {
	repo   Repository
	db     *database.DB
	tokens *token.Manager
	mailer Mailer
	cfg    *config.Config
}

// New builds the iam service.
func New(repo Repository, db *database.DB, tokens *token.Manager, mailer Mailer, cfg *config.Config) *Service {
	return &Service{repo: repo, db: db, tokens: tokens, mailer: mailer, cfg: cfg}
}
