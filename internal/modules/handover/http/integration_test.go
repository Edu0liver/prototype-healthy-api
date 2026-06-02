package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	conversationrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/infra/repository"
	conversationservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/conversation/service"
	handoverhttp "github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/http"
	handoverrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/infra/repository"
	handoverservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/handover/service"
	iamrepo "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/infra/repository"
	iamservice "github.com/Edu0liver/prototype-healthy-api/internal/modules/iam/service"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/channeladapter"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/config"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/database"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/events"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/middleware"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/redisx"
	"github.com/Edu0liver/prototype-healthy-api/internal/shared/testsupport"
	"github.com/Edu0liver/prototype-healthy-api/pkg/crypto"
	"github.com/Edu0liver/prototype-healthy-api/pkg/evolution"
	"github.com/Edu0liver/prototype-healthy-api/pkg/token"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type noopMailer struct{}

func (noopMailer) Send(string, string, string) error { return nil }

// noopEvo satisfies evolution.Client without making any network calls.
type noopEvo struct{}

func (noopEvo) CreateInstance(_ context.Context, _ evolution.CreateInstanceRequest) (*evolution.CreateInstanceResult, error) {
	return &evolution.CreateInstanceResult{}, nil
}
func (noopEvo) Connect(_ context.Context, _, _ string) (*evolution.ConnectResult, error) {
	return &evolution.ConnectResult{}, nil
}
func (noopEvo) ConnectionState(_ context.Context, _ string) (string, error) { return "open", nil }
func (noopEvo) Restart(_ context.Context, _ string) error                   { return nil }
func (noopEvo) Logout(_ context.Context, _ string) error                    { return nil }
func (noopEvo) DeleteInstance(_ context.Context, _ string) error            { return nil }
func (noopEvo) SendText(_ context.Context, _, _ string, _ evolution.SendTextRequest) (*evolution.SendResult, error) {
	return &evolution.SendResult{}, nil
}
func (noopEvo) SendPresence(_ context.Context, _, _, _, _ string) error { return nil }
func (noopEvo) MarkAsRead(_ context.Context, _, _, _, _ string) error   { return nil }
func (noopEvo) GetMediaBase64(_ context.Context, _, _, _ string) (string, string, error) {
	return "", "", nil
}

func seedCompany(t *testing.T, db *database.DB, slug string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	err := db.System(context.Background(), func(ctx context.Context) error {
		return database.MustTx(ctx).Exec(
			"INSERT INTO companies (id, name, slug, status, plan) VALUES (?, ?, ?, 'active', 'free')",
			id, slug, slug,
		).Error
	})
	require.NoError(t, err)
	return id
}

// newRouter returns (router, adminToken).
// rdb is nil — the handover service only touches Redis after a successful
// conversation lookup; all tests exercise the 401 or 404 path, so Redis
// is never reached.
func newRouter(t *testing.T, db *database.DB) (*gin.Engine, string) {
	t.Helper()
	cfg := &config.Config{}
	cfg.App.PublicBaseURL = "http://test"
	cfg.JWT.Secret = "test-secret-please-change"
	cfg.JWT.AccessTTL = 15 * time.Minute
	cfg.JWT.RefreshTTL = time.Hour

	tok := token.New(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
	slug := "ho-" + uuid.New().String()[:8]
	companyID := seedCompany(t, db, slug)

	ctx := context.Background()
	iamSvc := iamservice.New(iamrepo.New(), db, tok, noopMailer{}, cfg, nil)
	_, err := iamSvc.RegisterFirstAdmin(ctx, companyID, "admin@handover.test", "secret123", "Admin")
	require.NoError(t, err)
	tokens, _, err := iamSvc.Login(ctx, "admin@handover.test", "secret123")
	require.NoError(t, err)

	var rdb *redisx.Client
	mw := middleware.New(cfg, db, rdb, tok, zap.NewNop())

	pub := events.New(nil, zap.NewNop())
	convSvc := conversationservice.New(conversationrepo.New(), pub)

	cipher, err := crypto.New("0000000000000000000000000000000000000000000000000000000000000000")
	require.NoError(t, err)

	adapters := channeladapter.NewRegistry(noopEvo{})
	svc := handoverservice.New(convSvc, rdb, handoverrepo.New(), cipher, adapters)
	h := handoverhttp.NewHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	handoverhttp.RegisterRoutes(r.Group(""), h, mw)
	return r, tokens.Access
}

func do(t *testing.T, r *gin.Engine, method, path string, body any, bearer string) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestHandoverRoutes covers auth enforcement and not-found behavior for all
// /conversations/:id/handover/* endpoints.
// Full happy-path tests require an active conversation seeded by the webhook
// pipeline and are covered by E2E tests.
func TestHandoverRoutes(t *testing.T) {
	db := testsupport.NewPostgres(t)
	r, adminTok := newRouter(t, db)

	fakeID := uuid.New().String()
	badID := "not-a-uuid"

	type ep struct {
		method string
		path   string
		body   any // reply requires {content}; others accept empty body
	}
	endpoints := []ep{
		{http.MethodPost, "/conversations/" + fakeID + "/handover/take", nil},
		{http.MethodPost, "/conversations/" + fakeID + "/handover/reply", gin.H{"content": "hello"}},
		{http.MethodPost, "/conversations/" + fakeID + "/handover/return", nil},
		{http.MethodPost, "/conversations/" + fakeID + "/handover/close", nil},
	}

	for _, e := range endpoints {
		// Unauthenticated → 401.
		w := do(t, r, e.method, e.path, nil, "")
		require.Equal(t, http.StatusUnauthorized, w.Code, "%s noauth: %s", e.path, w.Body.String())

		// Authenticated + invalid UUID → 400.
		badPath := "/conversations/" + badID + "/handover/" + lastSegment(e.path)
		w = do(t, r, e.method, badPath, e.body, adminTok)
		require.Equal(t, http.StatusBadRequest, w.Code, "%s bad-id: %s", e.path, w.Body.String())

		// Authenticated + valid UUID not in DB → 404.
		w = do(t, r, e.method, e.path, e.body, adminTok)
		require.Equal(t, http.StatusNotFound, w.Code, "%s notfound: %s", e.path, w.Body.String())
	}
}

// lastSegment returns the last path component (e.g. "take" from "/conversations/x/handover/take").
func lastSegment(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
