package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/appctx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// newInviteSvc wires a service with a real token manager + config so Invite can
// mint the invite token and build the acceptance link.
func newInviteSvc(repo Repository, mailer Mailer) *Service {
	cfg := &config.Config{}
	cfg.App.PublicBaseURL = "https://panel.test"
	tok := token.New("test-secret-please-change", 15*time.Minute, time.Hour)
	return New(repo, nil, tok, mailer, cfg)
}

func inviteCtx() context.Context {
	return appctx.With(context.Background(), appctx.Identity{CompanyID: uuid.New()})
}

func TestInvite_OK(t *testing.T) {
	roleID := uuid.New()
	var created *models.User
	repo := &mockRepo{
		findByEmailFn:    func(context.Context, string) (*models.User, error) { return nil, repository.ErrNotFound },
		findRoleByNameFn: func(_ context.Context, n string) (*models.SystemRole, error) { return &models.SystemRole{ID: roleID, Name: n}, nil },
		createFn:         func(_ context.Context, u *models.User) error { created = u; return nil },
	}
	mailer := &mockMailer{}
	svc := newInviteSvc(repo, mailer)

	out, err := svc.Invite(inviteCtx(), "New.User@X.com", "New User", "operator")
	require.NoError(t, err)
	require.Equal(t, "new.user@x.com", out.Email, "email must be lowercased")
	require.Equal(t, "invited", out.Status)
	require.Equal(t, roleID, out.RoleID)
	require.Same(t, created, out)

	require.Len(t, mailer.sent, 1)
	require.Equal(t, "new.user@x.com", mailer.sent[0].to)
	require.True(t, strings.Contains(mailer.sent[0].html, "https://panel.test/accept-invite?token="), "email must carry the invite link")
}

func TestInvite_EmailTaken(t *testing.T) {
	repo := &mockRepo{findByEmailFn: func(_ context.Context, e string) (*models.User, error) {
		return &models.User{Email: e}, nil // found => taken
	}}
	svc := newInviteSvc(repo, &mockMailer{})
	out, err := svc.Invite(inviteCtx(), "dup@x.com", "Dup", "operator")
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrEmailTaken)
}

func TestInvite_InvalidRole(t *testing.T) {
	repo := &mockRepo{
		findByEmailFn:    func(context.Context, string) (*models.User, error) { return nil, repository.ErrNotFound },
		findRoleByNameFn: func(context.Context, string) (*models.SystemRole, error) { return nil, repository.ErrNotFound },
	}
	svc := newInviteSvc(repo, &mockMailer{})
	out, err := svc.Invite(inviteCtx(), "x@x.com", "X", "no-such-role")
	require.Nil(t, out)
	require.ErrorIs(t, err, ErrInvalidRole)
}

func TestInvite_FindByEmailError(t *testing.T) {
	boom := errors.New("db down")
	repo := &mockRepo{findByEmailFn: func(context.Context, string) (*models.User, error) { return nil, boom }}
	svc := newInviteSvc(repo, &mockMailer{})
	out, err := svc.Invite(inviteCtx(), "x@x.com", "X", "operator")
	require.Nil(t, out)
	require.ErrorIs(t, err, boom)
}

func TestInvite_MailerError(t *testing.T) {
	repo := &mockRepo{
		findByEmailFn:    func(context.Context, string) (*models.User, error) { return nil, repository.ErrNotFound },
		findRoleByNameFn: func(_ context.Context, n string) (*models.SystemRole, error) { return &models.SystemRole{ID: uuid.New(), Name: n}, nil },
	}
	mailer := &mockMailer{err: errors.New("smtp down")}
	svc := newInviteSvc(repo, mailer)
	out, err := svc.Invite(inviteCtx(), "x@x.com", "X", "operator")
	require.Nil(t, out)
	require.Error(t, err)
}
