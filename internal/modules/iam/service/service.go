// Package service holds the iam use cases (one file per use case: auth,
// invites, user management).
package service

import (
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"github.com/google/uuid"
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

// Tokens is an issued access/refresh pair.
type Tokens struct {
	Access  string
	Refresh string
}

// issueTokens mints a fresh access/refresh pair (shared by Login and Refresh).
func (s *Service) issueTokens(companyID, userID uuid.UUID, role string) (*Tokens, error) {
	access, err := s.tokens.GenerateAccess(companyID, userID, role)
	if err != nil {
		return nil, err
	}
	refresh, err := s.tokens.GenerateRefresh(companyID, userID)
	if err != nil {
		return nil, err
	}
	return &Tokens{Access: access, Refresh: refresh}, nil
}

func mustUUIDv7() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}
