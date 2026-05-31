package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"github.com/google/uuid"
)

// AcceptInvite validates an invite token and sets the user's password (public).
func (s *Service) AcceptInvite(ctx context.Context, inviteToken, password string) error {
	claims, err := s.tokens.Parse(inviteToken)
	if err != nil || claims.Type != token.TypeInvite {
		return ErrInvalidInvite
	}
	companyID, err := uuid.Parse(claims.CompanyID)
	if err != nil {
		return ErrInvalidInvite
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return ErrInvalidInvite
	}
	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	return s.db.Tenant(ctx, companyID, func(ctx context.Context) error {
		user, err := s.repo.FindByID(ctx, userID)
		if err != nil {
			return ErrInvalidInvite
		}
		if user.Status != "invited" {
			return ErrInvalidInvite
		}
		user.PasswordHash = hash
		user.Status = "active"
		return s.repo.Update(ctx, user)
	})
}
