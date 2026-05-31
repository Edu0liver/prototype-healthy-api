package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"github.com/google/uuid"
)

// inviteTTL is how long an invite link is valid.
const inviteTTL = 72 * time.Hour

// Invite creates an invited user and emails them an acceptance link. Runs in the
// caller's tenant transaction (admin only).
func (s *Service) Invite(ctx context.Context, email, name, roleName string) (*models.User, error) {
	companyID := appctx.CompanyID(ctx)
	email = strings.ToLower(email)

	if _, err := s.repo.FindByEmail(ctx, email); err == nil {
		return nil, ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	role, err := s.repo.FindRoleByName(ctx, roleName)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidRole
		}
		return nil, err
	}

	user := &models.User{
		ID:        mustUUIDv7(),
		CompanyID: companyID,
		Email:     email,
		Name:      name,
		RoleID:    role.ID,
		Status:    "invited",
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	inviteToken, err := s.tokens.GenerateInvite(companyID, user.ID, inviteTTL)
	if err != nil {
		return nil, err
	}
	link := fmt.Sprintf("%s/accept-invite?token=%s", s.cfg.App.PublicBaseURL, inviteToken)
	body := fmt.Sprintf(`<p>You have been invited.</p><p><a href="%s">Accept your invitation</a></p>`, link)
	if err := s.mailer.Send(email, "You're invited", body); err != nil {
		return nil, err
	}
	return user, nil
}

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

func mustUUIDv7() uuid.UUID {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New()
	}
	return id
}
