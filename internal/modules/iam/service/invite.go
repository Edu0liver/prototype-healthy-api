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
)

// inviteTTL is how long an invite link is valid.
const inviteTTL = 72 * time.Hour

// Invite creates an invited user and emails them an acceptance link. Runs in the
// caller's tenant transaction (admin only).
func (s *Service) Invite(ctx context.Context, email, name, roleName string) (*models.User, error) {
	companyID := appctx.CompanyID(ctx)
	email = strings.ToLower(email)

	if err := s.bill.EnsureResource(ctx, companyID, "seats"); err != nil {
		return nil, err
	}

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
