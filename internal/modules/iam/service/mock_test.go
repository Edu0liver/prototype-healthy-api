package service

import (
	"context"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/models"
	"github.com/google/uuid"
)

// mockRepo is a function-backed Repository for the service unit tests. Use cases
// wrapped in db.System/db.Tenant (Login, Refresh, RegisterFirstAdmin,
// AcceptInvite) need a real DB and are exercised by the http integration test.
type mockRepo struct {
	findByEmailGlobalFn func(ctx context.Context, email string) (uuid.UUID, uuid.UUID, error)
	findRoleByNameFn    func(ctx context.Context, name string) (*models.SystemRole, error)
	createFn            func(ctx context.Context, u *models.User) error
	updateFn            func(ctx context.Context, u *models.User) error
	findByEmailFn       func(ctx context.Context, email string) (*models.User, error)
	findByIDFn          func(ctx context.Context, id uuid.UUID) (*models.User, error)
	countUsersFn        func(ctx context.Context) (int64, error)
	listByCompanyFn     func(ctx context.Context) ([]models.User, error)
}

func (m *mockRepo) FindByEmailGlobal(ctx context.Context, email string) (uuid.UUID, uuid.UUID, error) {
	if m.findByEmailGlobalFn != nil {
		return m.findByEmailGlobalFn(ctx, email)
	}
	return uuid.Nil, uuid.Nil, nil
}

func (m *mockRepo) FindRoleByName(ctx context.Context, name string) (*models.SystemRole, error) {
	if m.findRoleByNameFn != nil {
		return m.findRoleByNameFn(ctx, name)
	}
	return &models.SystemRole{ID: uuid.New(), Name: name}, nil
}

func (m *mockRepo) Create(ctx context.Context, u *models.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, u)
	}
	return nil
}

func (m *mockRepo) Update(ctx context.Context, u *models.User) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, u)
	}
	return nil
}

func (m *mockRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.findByEmailFn != nil {
		return m.findByEmailFn(ctx, email)
	}
	return &models.User{Email: email}, nil
}

func (m *mockRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.findByIDFn != nil {
		return m.findByIDFn(ctx, id)
	}
	return &models.User{ID: id}, nil
}

func (m *mockRepo) CountUsers(ctx context.Context) (int64, error) {
	if m.countUsersFn != nil {
		return m.countUsersFn(ctx)
	}
	return 0, nil
}

func (m *mockRepo) ListByCompany(ctx context.Context) ([]models.User, error) {
	if m.listByCompanyFn != nil {
		return m.listByCompanyFn(ctx)
	}
	return nil, nil
}

// mockMailer records the last message sent (or fails on demand).
type mockMailer struct {
	sent []sentMail
	err  error
}

type sentMail struct{ to, subject, html string }

func (m *mockMailer) Send(to, subject, html string) error {
	if m.err != nil {
		return m.err
	}
	m.sent = append(m.sent, sentMail{to, subject, html})
	return nil
}
