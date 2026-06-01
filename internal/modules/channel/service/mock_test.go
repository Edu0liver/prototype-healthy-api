package service

import (
	"context"
	"testing"

	"github.com/Edu0liver/prototype-healthy-api/internal/modules/channel/infra/models"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/pkg/crypto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/evolution"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockRepo is a function-backed channel Repository.
type mockRepo struct {
	createFn func(ctx context.Context, c *models.Channel) error
	updateFn func(ctx context.Context, c *models.Channel) error
	getFn    func(ctx context.Context, id uuid.UUID) (*models.Channel, error)
	listFn   func(ctx context.Context) ([]models.Channel, error)
}

func (m *mockRepo) Create(ctx context.Context, c *models.Channel) error {
	if m.createFn != nil {
		return m.createFn(ctx, c)
	}
	return nil
}

func (m *mockRepo) Update(ctx context.Context, c *models.Channel) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, c)
	}
	return nil
}

func (m *mockRepo) Get(ctx context.Context, id uuid.UUID) (*models.Channel, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return &models.Channel{ID: id}, nil
}

func (m *mockRepo) List(ctx context.Context) ([]models.Channel, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}

// mockEvo is a function-backed evolution.Client; only the methods a test needs
// are wired, the rest return zero values.
type mockEvo struct {
	createInstanceFn  func(ctx context.Context, req evolution.CreateInstanceRequest) (*evolution.CreateInstanceResult, error)
	connectFn         func(ctx context.Context, instance, number string) (*evolution.ConnectResult, error)
	connectionStateFn func(ctx context.Context, instance string) (string, error)
	restartFn         func(ctx context.Context, instance string) error
	logoutFn          func(ctx context.Context, instance string) error
	deleteInstanceFn  func(ctx context.Context, instance string) error
}

func (m *mockEvo) CreateInstance(ctx context.Context, req evolution.CreateInstanceRequest) (*evolution.CreateInstanceResult, error) {
	if m.createInstanceFn != nil {
		return m.createInstanceFn(ctx, req)
	}
	return &evolution.CreateInstanceResult{}, nil
}

func (m *mockEvo) Connect(ctx context.Context, instance, number string) (*evolution.ConnectResult, error) {
	if m.connectFn != nil {
		return m.connectFn(ctx, instance, number)
	}
	return &evolution.ConnectResult{}, nil
}

func (m *mockEvo) ConnectionState(ctx context.Context, instance string) (string, error) {
	if m.connectionStateFn != nil {
		return m.connectionStateFn(ctx, instance)
	}
	return "", nil
}

func (m *mockEvo) Restart(ctx context.Context, instance string) error {
	if m.restartFn != nil {
		return m.restartFn(ctx, instance)
	}
	return nil
}

func (m *mockEvo) Logout(ctx context.Context, instance string) error {
	if m.logoutFn != nil {
		return m.logoutFn(ctx, instance)
	}
	return nil
}

func (m *mockEvo) DeleteInstance(ctx context.Context, instance string) error {
	if m.deleteInstanceFn != nil {
		return m.deleteInstanceFn(ctx, instance)
	}
	return nil
}

func (m *mockEvo) SendText(context.Context, string, string, evolution.SendTextRequest) (*evolution.SendResult, error) {
	return &evolution.SendResult{}, nil
}
func (m *mockEvo) SendPresence(context.Context, string, string, string, string) error { return nil }
func (m *mockEvo) MarkAsRead(context.Context, string, string, string, string) error   { return nil }
func (m *mockEvo) GetMediaBase64(context.Context, string, string, string) (string, string, error) {
	return "", "", nil
}

// testKeyHex is a 32-byte (AES-256) key in hex for the test cipher.
const testKeyHex = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

func newSvc(t *testing.T, repo Repository, evo evolution.Client) *Service {
	t.Helper()
	cipher, err := crypto.New(testKeyHex)
	require.NoError(t, err)
	return New(repo, evo, cipher, &config.Config{}, zap.NewNop())
}
